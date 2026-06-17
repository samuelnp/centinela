package gates

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

func skipWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-command tests assume a POSIX sh")
	}
}

// TestRunCustomExitZero: a successful command returns code 0, no timeout.
func TestRunCustomExitZero(t *testing.T) {
	skipWindows(t)
	out, code, timedOut := runCustom("true", time.Second, nil)
	if code != 0 || timedOut {
		t.Fatalf("true: code=%d timedOut=%v out=%q", code, timedOut, out)
	}
}

// TestRunCustomNonZero: false and an explicit non-zero exit both surface code != 0.
func TestRunCustomNonZero(t *testing.T) {
	skipWindows(t)
	if _, code, _ := runCustom("false", time.Second, nil); code == 0 {
		t.Fatal("false should be non-zero")
	}
	if _, code, _ := runCustom("exit 3", time.Second, nil); code != 3 {
		t.Fatalf("exit 3: code=%d, want 3", code)
	}
}

// TestRunCustomCapturesOutput: combined stdout/stderr is captured (trimmed).
func TestRunCustomCapturesOutput(t *testing.T) {
	skipWindows(t)
	out, _, _ := runCustom("printf 'a\\nb'", time.Second, nil)
	if out != "a\nb" {
		t.Fatalf("output = %q, want %q", out, "a\nb")
	}
}

// TestRunCustomTimeout: a long sleep under a tiny timeout returns fast as timedOut.
func TestRunCustomTimeout(t *testing.T) {
	skipWindows(t)
	start := time.Now()
	_, _, timedOut := runCustom("sleep 5", 50*time.Millisecond, nil)
	if !timedOut {
		t.Fatal("sleep should have timed out")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("timeout did not interrupt quickly")
	}
}

// TestRunCustomChangedFilesEnv: a non-empty changed list is injected as
// CENTINELA_CHANGED_FILES (newline-joined) and is visible to the command.
func TestRunCustomChangedFilesEnv(t *testing.T) {
	skipWindows(t)
	out, code, _ := runCustom("echo \"$CENTINELA_CHANGED_FILES\"", time.Second, []string{"a.go", "b.go"})
	if code != 0 {
		t.Fatalf("echo failed: code=%d out=%q", code, out)
	}
	if !strings.Contains(out, "a.go") || !strings.Contains(out, "b.go") {
		t.Fatalf("changed-files env not visible: %q", out)
	}
}

// TestRunCustomNoEnvWhenEmpty: an empty changed list leaves the env var unset.
func TestRunCustomNoEnvWhenEmpty(t *testing.T) {
	skipWindows(t)
	out, _, _ := runCustom("echo \"[$CENTINELA_CHANGED_FILES]\"", time.Second, nil)
	if strings.TrimSpace(out) != "[]" {
		t.Fatalf("env should be unset, got %q", out)
	}
}
