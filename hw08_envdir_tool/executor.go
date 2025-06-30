package main

import (
	"errors"
	"os"
	"os/exec"
)

func RunCmd(cmd []string, env Environment) (returnCode int) {
	for k, v := range env {
		if v.NeedRemove {
			_ = os.Unsetenv(k)
		} else {
			_ = os.Setenv(k, v.Value)
		}
	}

	executable := exec.Command(cmd[0], cmd[1:]...) //nolint
	executable.Stdout = os.Stdout
	executable.Stderr = os.Stderr
	executable.Stdin = os.Stdin

	err := executable.Run()
	if err == nil {
		return 0
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}
