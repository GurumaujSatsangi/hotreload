package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	logInfo("Startup")

	var root string
	var build string
	var exec string

	flag.StringVar(&root, "root", ".", "")
	flag.StringVar(&build, "build", "", "")
	flag.StringVar(&exec, "exec", "", "")
	flag.Parse()

	cfg := config.NewConfig(root, build, exec)
	logInfo("Configuration loaded")

	w, err := watcher.NewWatcher(cfg.RootDir)
	if err != nil {
		logError(fmt.Sprintf("Watcher initialization failed: %v", err))
		os.Exit(1)
	}
	logInfo("Watcher started")

	d := debounce.NewDebouncer(w.Events())
	b := builder.NewBuilder(cfg)
	r := runner.NewRunner(cfg)

	appCtx, stopApp := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopApp()

	results := make(chan buildResult, 16)

	buildID := 0
	latestID := 0
	var cancelBuild context.CancelFunc
	serverRunning := false

	startBuild := func(reason string) {
		if cancelBuild != nil {
			cancelBuild()
		}

		buildID++
		id := buildID
		latestID = id

		ctx, cancel := context.WithCancel(appCtx)
		cancelBuild = cancel

		logInfo("Build started")

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
			if serverRunning {
				if err := r.Stop(); err != nil {
					logError(fmt.Sprintf("Server stop failed: %v", err))
				} else {
					serverRunning = false
					logInfo("Server stopped")
				}
			}
			logInfo("Shutdown complete")
			return

		case _, ok := <-d.Output():
			if !ok {
				if cancelBuild != nil {
					cancelBuild()
				}
				if serverRunning {
					if err := r.Stop(); err != nil {
						logError(fmt.Sprintf("Server stop failed: %v", err))
					} else {
						serverRunning = false
						logInfo("Server stopped")
					}
				}
				logInfo("Watch pipeline closed")
				return
			}

			logInfo("File change detected")
			startBuild("change")

		case result := <-results:
			if result.id != latestID {
				continue
			}

			if result.err != nil {
				if errors.Is(result.err, context.Canceled) {
					logInfo("Build canceled")
					continue
				}
				logError(fmt.Sprintf("Build failed: %v", result.err))
				continue
			}

			logInfo("Build succeeded")

			if !serverRunning {
				if err := r.Start(); err != nil {
					logError(fmt.Sprintf("Server start failed: %v", err))
					continue
				}
				serverRunning = true
				logInfo("Server started")
				continue
			}

			if err := r.Stop(); err != nil {
				logError(fmt.Sprintf("Server stop failed: %v", err))
				continue
			}
			serverRunning = false
			logInfo("Server stopped")

			if err := r.Start(); err != nil {
				logError(fmt.Sprintf("Server start failed: %v", err))
				continue
			}
			serverRunning = true
			logInfo("Server restarted")
		}
	}
}

func logInfo(message string) {
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), message)
}

func logError(message string) {
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), message)
}
