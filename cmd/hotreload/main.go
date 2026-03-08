package main

import (
	"flag"
	"log/slog"

	"github.com/GurumaujSatsangi/hotreload/internal/config"
)

func main() {
	var root string
	var build string
	var exec string

	flag.StringVar(&root, "root", ".", "")
	flag.StringVar(&build, "build", "", "")
	flag.StringVar(&exec, "exec", "", "")
	flag.Parse()

	cfg := config.NewConfig(root, build, exec)
	slog.Info("parsed config", "root", cfg.RootDir, "build", cfg.BuildCmd, "exec", cfg.ExecCmd)
}
