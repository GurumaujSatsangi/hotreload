package runner

import (
	"errors"
	"os"
	"os/exec"
)

func configureProcessGroup(cmd *exec.Cmd) {
}

func killProcessGroup(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}

	return nil
}
