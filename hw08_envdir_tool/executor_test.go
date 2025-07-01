package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	t.Run("simple test", func(t *testing.T) {
		testEnv := Environment{
			"BAR": EnvValue{"bar", false},
		}
		os.Setenv("ADDED", "original")
		cmd := []string{"sh", "-c", "echo -n $BAR $ADDED"}

		// forwarding stdout for capturing
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// executing command
		exitCode := RunCmd(cmd, testEnv)

		w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)

		// taking stdout back to normal
		os.Stdout = oldStdout
		capturedOutput := buf.String()
		require.Equal(t, "bar original", capturedOutput)
		require.Equal(t, 0, exitCode)
	})

	t.Run("exit code simple test", func(t *testing.T) {
		testEnv := Environment{
			"BAR": EnvValue{"bar", false},
		}
		cmd := []string{"sh", "-c", "exit 123"}
		exitCode := RunCmd(cmd, testEnv)
		require.Equal(t, 123, exitCode)
	})
}
