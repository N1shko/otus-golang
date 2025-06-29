package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var tmpFileName = "out.tmp"

func TestOffsetExceed(t *testing.T) {
	t.Run("test offset exceed", func(t *testing.T) {
		testout, err := os.Create(tmpFileName)
		if err != nil {
			fmt.Print(err)
			return
		}
		defer func() {
			testout.Close()
			os.Remove(tmpFileName)
		}()
		err = Copy("testdata/input.txt", tmpFileName, 1000000, 0)
		require.Error(t, err, ErrOffsetExceedsFileSize)
	})
}

func TestInputFileNotExist(t *testing.T) {
	t.Run("test offset exceed", func(t *testing.T) {
		testout, err := os.Create(tmpFileName)
		if err != nil {
			fmt.Print(err)
			return
		}
		defer func() {
			testout.Close()
			os.Remove(tmpFileName)
		}()
		err = Copy("testdata/input1.txt", tmpFileName, 1000000, 0)
		require.Error(t, err, os.ErrNotExist)
	})
}

func TestFromShellScript(t *testing.T) {
	t.Run("test offset exceed", func(t *testing.T) {
		testout, err := os.Create(tmpFileName)
		if err != nil {
			fmt.Print(err)
			return
		}
		defer func() {
			testout.Close()
			os.Remove(tmpFileName)
		}()
		err = Copy("testdata/input.txt", tmpFileName, 6000, 1000)
		require.Nil(t, err, nil)
		sample, err := os.ReadFile("testdata/out_offset6000_limit1000.txt")
		if err != nil {
			fmt.Print(err)
			return
		}
		res, err := os.ReadFile(tmpFileName)
		if err != nil {
			fmt.Print(err)
			return
		}
		require.Equal(t, res, sample)
	})
}
