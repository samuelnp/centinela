package gates

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestToolPresent_RealBinaryFound verifies that a real PATH binary is found.
func TestToolPresent_RealBinaryFound(t *testing.T) {
	if !toolPresent("go") {
		t.Fatal("go must be on PATH in test environment")
	}
}

// TestToolPresent_AbsentBinaryReturnsFalse verifies a made-up name returns false.
func TestToolPresent_AbsentBinaryReturnsFalse(t *testing.T) {
	if toolPresent("centinela-no-such-binary-xyz") {
		t.Fatal("non-existent tool must return false")
	}
}

// TestExitCode_ZeroOnNil confirms exitCode returns -1 for a nil error.
func TestExitCode_ZeroOnNil(t *testing.T) {
	if c := exitCode(nil); c != -1 {
		t.Fatalf("nil error should return -1, got %d", c)
	}
}

// TestExitCode_ReturnsExitStatus verifies an ExitError's code is extracted.
func TestExitCode_ReturnsExitStatus(t *testing.T) {
	cmd := exec.Command("sh", "-c", "exit 42")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 42")
	}
	if c := exitCode(err); c != 42 {
		t.Fatalf("expected exit code 42, got %d", c)
	}
}

// TestExitCode_NonExecError returns -1 for a non-ExitError.
func TestExitCode_NonExecError(t *testing.T) {
	if c := exitCode(os.ErrClosed); c != -1 {
		t.Fatalf("non-ExitError must return -1, got %d", c)
	}
}

// TestRunScanner_CapturesStdout verifies stdout capture works correctly.
func TestRunScanner_CapturesStdout(t *testing.T) {
	// Use a real binary that exists (echo via sh)
	d := t.TempDir()
	script := filepath.Join(d, "echo-hello")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho hello\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	stdout, _, runErr := runScanner(script)
	// Non-zero exit is OK; we just want stdout captured
	if runErr != nil && exitCode(runErr) < 0 {
		t.Fatalf("launcher error: %v", runErr)
	}
	if string(stdout) != "hello\n" {
		t.Fatalf("expected 'hello\\n', got %q", stdout)
	}
}

// TestRunScanner_CapturesBothStreams verifies both stdout and stderr captured.
func TestRunScanner_CapturesBothStreams(t *testing.T) {
	d := t.TempDir()
	script := filepath.Join(d, "mixed-output")
	body := "#!/bin/sh\necho out\necho err 1>&2\nexit 1\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, _ := runScanner(script)
	if string(stdout) != "out\n" {
		t.Fatalf("stdout: expected 'out\\n', got %q", stdout)
	}
	if string(stderr) != "err\n" {
		t.Fatalf("stderr: expected 'err\\n', got %q", stderr)
	}
}

// TestToolPresent_EmptyPathReturnsFalse verifies absent tool with empty PATH.
func TestToolPresent_EmptyPathReturnsFalse(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir) // only empty tmp dir on PATH
	if toolPresent("gitleaks") {
		t.Fatal("gitleaks must not be found on empty PATH")
	}
}
