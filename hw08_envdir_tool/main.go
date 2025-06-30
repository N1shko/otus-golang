package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Not enough arguments \nUsage: go-envdir /path/to/env/dir command arg1 arg2")
		return
	}
	executable := os.Args[2]
	_, err := exec.LookPath(executable)
	if err != nil {
		fmt.Printf("Executable '%s' not found in PATH: %v\n", executable, err)
		return
	}

	envDir := os.Args[1]
	_, err = os.Stat(envDir)
	if err != nil {
		fmt.Printf("Dir %s does not exist\n", envDir)
		return
	}
	env, err := ReadDir(envDir)
	if err != nil {
		return
	}
	os.Exit(RunCmd(os.Args[2:], env))
}
