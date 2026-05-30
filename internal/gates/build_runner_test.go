package gates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// writeScript writes an executable shell script and returns its path.
func writeScript(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "script.sh")
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return p
}

func TestBuildTarget_Success(t *testing.T) {
	err := buildTarget("go version", config.BuildTarget{GOOS: "linux", GOARCH: "amd64"})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestBuildTarget_FailureNamesTarget(t *testing.T) {
	script := writeScript(t, "echo boom 1>&2\nexit 1\n")
	err := buildTarget(script, config.BuildTarget{GOOS: "windows", GOARCH: "arm64"})
	if err == nil {
		t.Fatal("expected failure")
	}
	if !strings.HasPrefix(err.Error(), "windows/arm64:") {
		t.Fatalf("error should name target, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("error should carry first stderr line, got %q", err.Error())
	}
}

func TestBuildTarget_EmptyCommand(t *testing.T) {
	err := buildTarget("   ", config.BuildTarget{GOOS: "linux", GOARCH: "amd64"})
	if err == nil || !strings.Contains(err.Error(), "empty build command") {
		t.Fatalf("expected empty-command error, got %v", err)
	}
}

func TestFirstStderrLine_FallsBackToRunErr(t *testing.T) {
	got := firstStderrLine("\n  \n", os.ErrClosed)
	if got != os.ErrClosed.Error() {
		t.Fatalf("expected fallback to run error, got %q", got)
	}
}

// targetSensitiveScript fails only for a single GOOS/GOARCH pair.
func targetSensitiveScript(t *testing.T, goos, goarch string) string {
	return writeScript(t,
		`if [ "$GOOS" = `+goos+` ] && [ "$GOARCH" = `+goarch+` ]; then exit 1; fi
exit 0
`)
}

func TestRunTargets_AggregatesFailures(t *testing.T) {
	script := targetSensitiveScript(t, "windows", "arm64")
	targets := []config.BuildTarget{
		{GOOS: "linux", GOARCH: "amd64"},
		{GOOS: "windows", GOARCH: "arm64"},
		{GOOS: "darwin", GOARCH: "amd64"},
	}
	failures := runTargets(script, targets)
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(failures))
	}
	if failures[0].Target.GOOS != "windows" {
		t.Fatalf("expected windows failure, got %+v", failures[0].Target)
	}
}

func TestRunTargets_AllPass(t *testing.T) {
	failures := runTargets("go version", []config.BuildTarget{
		{GOOS: "linux", GOARCH: "amd64"},
		{GOOS: "darwin", GOARCH: "arm64"},
	})
	if len(failures) != 0 {
		t.Fatalf("expected no failures, got %d: %+v", len(failures), failures)
	}
}
