package runner

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/GurumaujSatsangi/hotreload/internal/config"
)

type Runner struct {
	cfg config.Config
	cmd *exec.Cmd
	mu  sync.Mutex
}

func NewRunner(cfg config.Config) *Runner {
	return &Runner{cfg: cfg}
}

func (r *Runner) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cmd != nil && r.cmd.Process != nil {
		if r.cmd.ProcessState == nil || !r.cmd.ProcessState.Exited() {
			return errors.New("process already running")
		}
	}

	cmd := commandForShell(r.cfg.ExecCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = r.cfg.RootDir

	if err := cmd.Start(); err != nil {
		return err
	}

	r.cmd = cmd
	return nil
}

func (r *Runner) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}

	if r.cmd.ProcessState != nil && r.cmd.ProcessState.Exited() {
		r.cmd = nil
		return nil
	}

	if err := r.cmd.Process.Kill(); err != nil {
		return err
	}

	if err := r.cmd.Wait(); err != nil {
		r.cmd = nil
		return err
	}

	r.cmd = nil
	return nil
}

func commandForShell(command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/C", command)
	}
	return exec.Command("sh", "-c", command)
}
