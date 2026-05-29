package verify

import (
	"strings"
	"testing"
	"time"
)

func TestExecRunner(t *testing.T) {
	r := NewExecRunner()

	ok := r.Run("", "printf hello", 5*time.Second)
	if ok.ExitCode != 0 || !strings.Contains(ok.Output, "hello") {
		t.Fatalf("exit0 run = %+v", ok)
	}

	fail := r.Run("", "exit 3", 5*time.Second)
	if fail.ExitCode != 3 {
		t.Fatalf("nonzero run exit = %d want 3", fail.ExitCode)
	}
}

func TestExecRunnerTimeout(t *testing.T) {
	out := NewExecRunner().Run("", "sleep 2", 50*time.Millisecond)
	if !out.TimedOut {
		t.Fatalf("expected timeout, got %+v", out)
	}
}
