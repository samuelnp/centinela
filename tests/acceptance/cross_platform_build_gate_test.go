package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// Acceptance: specs/cross-platform-build-gate.feature
//
// Simulates the matrix with synthetic per-target scripts (no real
// cross-compiles). The broken target is expressed as an explicit script file
// that exits non-zero for one GOOS/GOARCH — never a shell-conditional in the
// command field, because buildTarget argv-parses and does not invoke a shell.

func sixTargets() []config.BuildTarget {
	return []config.BuildTarget{
		{GOOS: "linux", GOARCH: "amd64"},
		{GOOS: "linux", GOARCH: "arm64"},
		{GOOS: "darwin", GOARCH: "amd64"},
		{GOOS: "darwin", GOARCH: "arm64"},
		{GOOS: "windows", GOARCH: "amd64"},
		{GOOS: "windows", GOARCH: "arm64"},
	}
}

func writeGateScript(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "build.sh")
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return p
}

func buildResult(command string, targets []config.BuildTarget) gates.Result {
	cfg := &config.Config{Gates: config.GatesConfig{Build: config.BuildGateConfig{
		Enabled: true, Command: command, Targets: targets,
	}}}
	for _, r := range gates.RunWithFilter(cfg, nil) {
		if r.Name == "G-Build: Cross-Compile" {
			return r
		}
	}
	return gates.Result{}
}

func TestBuildGate_AllTargetsCompile_Passes(t *testing.T) {
	r := buildResult("go version", sixTargets())
	if r.Status != gates.Pass {
		t.Fatalf("expected Pass, got %v (%q)", r.Status, r.Message)
	}
	if r.Message != "All 6 release targets compile." {
		t.Fatalf("unexpected pass message: %q", r.Message)
	}
}

func TestBuildGate_OneTargetFails_NamesGOOSGOARCH(t *testing.T) {
	// Explicit script: fail only for windows/amd64.
	script := writeGateScript(t,
		`if [ "$GOOS" = windows ] && [ "$GOARCH" = amd64 ]; then echo undefined: syscall.Flock 1>&2; exit 1; fi
exit 0
`)
	r := buildResult(script, sixTargets())
	if r.Status != gates.Fail {
		t.Fatalf("expected Fail, got %v", r.Status)
	}
	joined := strings.Join(r.Details, "\n")
	if !strings.Contains(joined, "windows/amd64") {
		t.Fatalf("details should name windows/amd64, got %q", joined)
	}
	for _, other := range []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/arm64"} {
		if strings.Contains(joined, other) {
			t.Fatalf("only windows/amd64 should fail, but details mention %s: %q", other, joined)
		}
	}
}
