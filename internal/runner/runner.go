package runner

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/GurumaujSatsangi/hotreload/internal/config"
)

type Runner struct {
	cfg                config.Config
	cmd                *exec.Cmd
	done               chan struct{}
	mu                 sync.Mutex
	stopRequested      bool
	restartAllowedAfter time.Time
}

func NewRunner(cfg config.Config) *Runner {
	return &Runner{cfg: cfg}
}

func (r *Runner) Start() error {
	for {
		r.mu.Lock()

		if r.cmd != nil && r.cmd.Process != nil {
			r.mu.Unlock()
			return errors.New("process already running")
		}

		delay := time.Until(r.restartAllowedAfter)
		if delay > 0 {
			r.mu.Unlock()
			time.Sleep(delay)
			continue
		}

		cmd := commandForShell(r.cfg.ExecCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Dir = r.cfg.RootDir
		configureProcessGroup(cmd)

		if err := cmd.Start(); err != nil {
			r.mu.Unlock()
			return err
		}

		r.stopRequested = false
		r.cmd = cmd
		r.done = make(chan struct{})
		startedAt := time.Now()
		done := r.done
		r.mu.Unlock()

		go r.waitProcess(cmd, done, startedAt)
		return nil
	}
}

func (r *Runner) Stop() error {
	r.mu.Lock()
	if r.cmd == nil || r.cmd.Process == nil {
		r.mu.Unlock()
		return nil
	}

	r.stopRequested = true
	cmd := r.cmd
	done := r.done
	r.mu.Unlock()

	if err := killProcessGroup(cmd); err != nil {
		return err
	}

	if done != nil {
		<-done
	}

	return nil
}

func (r *Runner) waitProcess(cmd *exec.Cmd, done chan struct{}, startedAt time.Time) {
	_ = cmd.Wait()

	r.mu.Lock()
	runtime := time.Since(startedAt)
	stopRequested := r.stopRequested
	if r.cmd == cmd {
		r.cmd = nil
		r.done = nil
	}
	r.stopRequested = false
	if !stopRequested && runtime < 2*time.Second {
		r.restartAllowedAfter = time.Now().Add(3 * time.Second)
	}
	r.mu.Unlock()

	close(done)
}

func commandForShell(command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/C", command)
	}
	return exec.Command("sh", "-c", command)
}
