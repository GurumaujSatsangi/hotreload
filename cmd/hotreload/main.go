package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/GurumaujSatsangi/hotreload/internal/builder"

	"github.com/GurumaujSatsangi/hotreload/internal/config"
	"github.com/GurumaujSatsangi/hotreload/internal/debounce"
	"github.com/GurumaujSatsangi/hotreload/internal/runner"
	"github.com/GurumaujSatsangi/hotreload/internal/watcher"
)

type buildResult struct {
	id  int
	err error
}

func main() {

	println("program started")
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	var root string
	var build string
	var exec string

	flag.StringVar(&root, "root", ".", "")
	flag.StringVar(&build, "build", "", "")
	flag.StringVar(&exec, "exec", "", "")
	flag.Parse()

	cfg := config.NewConfig(root, build, exec)
	slog.Info("config loaded", "root", cfg.RootDir, "build", cfg.BuildCmd, "exec", cfg.ExecCmd)

	w, err := watcher.NewWatcher(cfg.RootDir)
	if err != nil {
		slog.Error("failed to create watcher", "error", err)
		os.Exit(1)
	}

	d := debounce.NewDebouncer(w.Events())
	b := builder.NewBuilder(cfg)
	r := runner.NewRunner(cfg)

	appCtx, stopApp := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopApp()

	results := make(chan buildResult, 16)

	buildID := 0
	latestID := 0
	var cancelBuild context.CancelFunc

	startBuild := func(reason string) {
		if cancelBuild != nil {
			cancelBuild()
		}

		buildID++
		id := buildID
		latestID = id

		ctx, cancel := context.WithCancel(appCtx)
		cancelBuild = cancel

		slog.Info("build started", "id", id, "reason", reason)

		go func(buildCtx context.Context, currentID int) {
			err := b.Build(buildCtx)
			results <- buildResult{id: currentID, err: err}
		}(ctx, id)
	}

	startBuild("startup")

	for {
		select {
		case <-appCtx.Done():
			if cancelBuild != nil {
				cancelBuild()
			}
			if err := r.Stop(); err != nil {
				slog.Error("failed to stop server", "error", err)
			}
			slog.Info("shutdown complete")
			return

		case _, ok := <-d.Output():
			if !ok {
				if cancelBuild != nil {
					cancelBuild()
				}
				if err := r.Stop(); err != nil {
					slog.Error("failed to stop server", "error", err)
				}
				slog.Info("watch pipeline closed")
				return
			}

			startBuild("change")

		case result := <-results:
			if result.id != latestID {
				slog.Info("discarded outdated build result", "id", result.id)
				continue
			}

			if result.err != nil {
				if errors.Is(result.err, context.Canceled) {
					slog.Info("build canceled", "id", result.id)
					continue
				}
				slog.Error("build failed", "id", result.id, "error", result.err)
				continue
			}

			slog.Info("build succeeded", "id", result.id)

			if err := r.Restart(); err != nil {
				slog.Error("failed to restart server", "error", err)
				continue
			}

			slog.Info("server restarted")
		}
	}
}
