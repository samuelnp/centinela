package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// doctorTempRepo chdirs into a fresh non-git temp repo with a .workflow/ dir.
func doctorTempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	t.Chdir(dir)
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestRunDoctorHealthyExitsZero(t *testing.T) {
	doctorTempRepo(t)
	p, _ := setup.BuildSyncPlan("both")
	if err := setup.ApplySync(p); err != nil {
		t.Fatal(err)
	}
	doctorFix = false
	out := captureStdout(t, func() {
		if err := runDoctor(nil, nil); err != nil {
			t.Errorf("healthy doctor must exit 0, got %v", err)
		}
	})
	if !strings.Contains(out, "ok,") {
		t.Fatalf("summary line missing: %q", out)
	}
}

func TestRunDoctorErrorExitsOne(t *testing.T) {
	doctorTempRepo(t)
	if err := os.MkdirAll(".claude", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".claude/settings.json", []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	doctorFix = false
	var runErr error
	captureStdout(t, func() { runErr = runDoctor(nil, nil) })
	if runErr == nil {
		t.Fatal("missing hooks must drive a non-nil error (exit 1)")
	}
}

func TestRunDoctorFixRepairsHooks(t *testing.T) {
	doctorTempRepo(t)
	if err := os.MkdirAll(".claude", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".claude/settings.json", []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	doctorFix = true
	t.Cleanup(func() { doctorFix = false })
	out := captureStdout(t, func() { _ = runDoctor(nil, nil) })
	if strings.Contains(out, "✗ hooks") {
		t.Fatalf("--fix should repair hooks, still showing error: %q", out)
	}
}
