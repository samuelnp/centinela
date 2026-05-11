package main

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
)

var runtimeOS = runtime.GOOS

// runCommand executes a shell command and returns (passed, combined output).
func runCommand(cmdStr string) (bool, string) {
	var cmd *exec.Cmd
	if runtimeOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	return err == nil, strings.TrimSpace(buf.String())
}
