package gates

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// runCustom executes one custom gate's command under a shell with a timeout,
// returning the combined stdout+stderr, the process exit code, and whether it
// timed out. It mirrors cmd/centinela/validate_runner.go's shell model (sh -c /
// cmd /C) wrapped with the security_exec.go timeout pattern.
//
// changed is the diff-aware file list; when non-empty it is injected as the
// newline-joined CENTINELA_CHANGED_FILES env var (empty => env var unset, so
// the command full-scans). A launch failure surfaces as a non-zero exit code
// (-1) with the shell error captured in output, never a panic.
func runCustom(command string, timeout time.Duration, changed []string) (output string, code int, timedOut bool) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	cmd.Env = os.Environ()
	if len(changed) > 0 {
		cmd.Env = append(cmd.Env, "CENTINELA_CHANGED_FILES="+strings.Join(changed, "\n"))
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	// On timeout, CommandContext kills the shell, but a grandchild (e.g. `sleep`)
	// can inherit the output pipe and keep Run blocked until it exits. WaitDelay
	// makes Wait force-close the pipes and return promptly after the kill, so a
	// hung command never stalls validate past the timeout.
	cmd.WaitDelay = time.Second

	err := cmd.Run()
	output = strings.TrimSpace(buf.String())
	if ctx.Err() == context.DeadlineExceeded {
		return output, exitCode(err), true
	}
	if err == nil {
		return output, 0, false
	}
	return output, exitCode(err), false
}
