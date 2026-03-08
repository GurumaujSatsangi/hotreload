package builder

import (
	"context"
	"os"
	"os/exec"
	"runtime"

	"github.com/GurumaujSatsangi/hotreload/internal/config"
)

type Builder struct {
	cfg config.Config
}

func NewBuilder(cfg config.Config) *Builder {
	return &Builder{cfg: cfg}
}

func (b *Builder) Build(ctx context.Context) error {
	cmd := commandForShell(ctx, b.cfg.BuildCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = b.cfg.RootDir
	return cmd.Run()
}

func commandForShell(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/C", command)
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
}
