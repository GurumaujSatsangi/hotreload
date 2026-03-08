package runner

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
	return r.startLocked()
}

func (r *Runner) Restart() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.stopLocked(); err != nil {
		return err
	}

	if err := r.startLocked(); err != nil {
		return err
	}

	return nil
}

func (r *Runner) startLocked() error {

	if r.cmd != nil && r.cmd.Process != nil {
		if r.cmd.ProcessState != nil && r.cmd.ProcessState.Exited() {
			_ = r.cmd.Wait()
			r.cmd = nil
		} else {
			return errors.New("process already running")
		}
	}

	cmd, err := commandForExec(r.cfg.ExecCmd)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = r.cfg.RootDir
	configureProcessGroup(cmd)

	if err := cmd.Start(); err != nil {
		return err
	}

	r.cmd = cmd
	return nil
}

func (r *Runner) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stopLocked()
}

func (r *Runner) stopLocked() error {

	if r.cmd == nil || r.cmd.Process == nil {
		r.cmd = nil
		return nil
	}

	cmd := r.cmd
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		_ = cmd.Wait()
		r.cmd = nil
		return nil
	}

	if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}

	if err := cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) && !errors.Is(err, os.ErrProcessDone) {
			return err
		}
	}

	r.cmd = nil
	return nil
}

func commandForExec(command string) (*exec.Cmd, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, errors.New("empty exec command")
	}

	parts[0] = normalizeExecutable(parts[0])
	return exec.Command(parts[0], parts[1:]...), nil
}

func normalizeExecutable(target string) string {
	if runtime.GOOS != "windows" {
		return target
	}
	if !isLocalExecutable(target) {
		return target
	}
	if filepath.Ext(target) != "" {
		return target
	}
	return target + ".exe"
}

func isLocalExecutable(target string) bool {
	if filepath.IsAbs(target) {
		return true
	}
	if strings.HasPrefix(target, ".") {
		return true
	}
	if strings.Contains(target, "/") || strings.Contains(target, "\\") {
		return true
	}
	return false
}
