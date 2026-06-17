package verify

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"runtime"
	"time"
)

// RunOutcome captures everything a claim check needs from one command run.
type RunOutcome struct {
	ExitCode int
	Output   string
	TimedOut bool
	StartErr error // command could not start (e.g. binary not on PATH)
}

// CommandRunner runs a shell command in dir, bounded by timeout. Injected so
// claim checks are unit-testable without shelling out.
type CommandRunner interface {
	Run(dir, command string, timeout time.Duration) RunOutcome
}

// execRunner is the default CommandRunner backed by os/exec.
type execRunner struct{}

// NewExecRunner returns the production CommandRunner.
func NewExecRunner() CommandRunner { return execRunner{} }

func (execRunner) Run(dir, command string, timeout time.Duration) RunOutcome {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	cmd.Dir = dir

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	out := buf.String()
	if ctx.Err() == context.DeadlineExceeded {
		return RunOutcome{Output: out, TimedOut: true, ExitCode: -1}
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return RunOutcome{ExitCode: exitErr.ExitCode(), Output: out}
	}
	if err != nil {
		return RunOutcome{StartErr: err, Output: out, ExitCode: -1}
	}
	return RunOutcome{ExitCode: 0, Output: out}
}
