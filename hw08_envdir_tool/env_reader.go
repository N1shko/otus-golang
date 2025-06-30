package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	env := make(Environment)
	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf("Found dir %s, skipping...\n", entry.Name())
			continue
		}
		if strings.Contains(entry.Name(), "=") {
			fmt.Printf("FileName %s contains forbidden characters, skipping\n", entry.Name())
			continue
		}
		filePath := dir + "/" + entry.Name()
		info, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("Error occurred while getting file %s stats: %s\n", entry.Name(), err)
			return nil, err
		}
		if info.Size() == 0 {
			env[entry.Name()] = EnvValue{"", true}
			continue
		}

		envBytes, err := func() ([]byte, error) {
			f, err := os.Open(filePath)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			reader := bufio.NewReader(f)
			line, err := reader.ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, err
			}

			return []byte(line), nil
		}()
		if err != nil {
			fmt.Printf("Error occurred while reading file %s: %s\n", entry.Name(), err)
			continue
		}
		envBytes = bytes.ReplaceAll(envBytes, []byte{0x00}, []byte("\n"))
		curEnv := strings.TrimRight(string(envBytes), " \t\n")
		env[entry.Name()] = EnvValue{curEnv, false}
	}
	return env, nil
}
