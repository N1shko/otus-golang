package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadDir(t *testing.T) {
	t.Run("shell script test", func(t *testing.T) {
		env, err := ReadDir("testdata/env")
		require.Nil(t, err)
		require.Equal(t, EnvValue{"\"hello\"", false}, env["HELLO"])
		require.Equal(t, EnvValue{"bar", false}, env["BAR"])
		require.Equal(t, EnvValue{"   foo\nwith new line", false}, env["FOO"])
		require.Equal(t, EnvValue{"", true}, env["UNSET"])
		require.Equal(t, EnvValue{"", false}, env["EMPTY"])
	})

	t.Run("empty dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		env, err := ReadDir(tmpDir)
		require.Nil(t, err)
		require.Equal(t, Environment{}, env)
	})
}
