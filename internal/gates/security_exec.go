package gates

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"
)

// scanTimeout bounds every external scanner invocation. A wedged scan degrades
// to a Warn (via errScanTimeout) rather than hanging centinela validate. The
// value is a fixed internal default in v1 (not user-configurable).
const scanTimeout = 120 * time.Second

// errScanTimeout signals that a scanner exceeded scanTimeout; callers map it to
// a Warn result (the scan ran but produced no usable verdict).
var errScanTimeout = errors.New("scanner timed out")

// toolPresent reports whether a scanner binary is resolvable on PATH. It is the
// single place the gate decides "tool absent -> Skip" before any scan runs.
func toolPresent(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

// runScanner executes tool with args under a fixed timeout, returning captured
// stdout and stderr plus the run error. A non-zero exit is NOT treated as a
// helper error: scanners (gitleaks, govulncheck) signal findings via exit code,
// so the caller inspects (stdout, runErr) together. A timeout maps to
// errScanTimeout so the caller can emit Warn instead of crashing or false-Pass.
func runScanner(tool string, args ...string) (stdout, stderr []byte, runErr error) {
	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, tool, args...)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return out.Bytes(), errBuf.Bytes(), errScanTimeout
	}
	return out.Bytes(), errBuf.Bytes(), err
}

// exitCode extracts the process exit code from a run error, or -1 when the
// error is not an *exec.ExitError (e.g. the binary could not be launched).
func exitCode(err error) int {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}
