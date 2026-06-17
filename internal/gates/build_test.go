package gates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func buildCfg(command string, targets ...config.BuildTarget) *config.Config {
	return &config.Config{Gates: config.GatesConfig{Build: config.BuildGateConfig{
		Enabled: true, Command: command, Targets: targets,
	}}}
}

func TestBuildEnv_SetsTargetAndCGO(t *testing.T) {
	env := buildEnv(config.BuildTarget{GOOS: "windows", GOARCH: "arm64"})
	want := map[string]bool{
		"GOOS=windows": false, "GOARCH=arm64": false, "CGO_ENABLED=0": false,
	}
	for _, e := range env {
		if _, ok := want[e]; ok {
			want[e] = true
		}
	}
	for k, seen := range want {
		if !seen {
			t.Fatalf("expected env to contain %q", k)
		}
	}
}

func TestCheckBuild_SkipWhenNoTargets(t *testing.T) {
	r := checkBuild(buildCfg("go version"))
	if r.Status != Skip {
		t.Fatalf("expected Skip, got %v", r.Status)
	}
	if !strings.Contains(r.Message, "no targets") {
		t.Fatalf("expected no-targets message, got %q", r.Message)
	}
}

func TestCheckBuild_PassWhenAllCompile(t *testing.T) {
	r := checkBuild(buildCfg("go version",
		config.BuildTarget{GOOS: "linux", GOARCH: "amd64"},
		config.BuildTarget{GOOS: "darwin", GOARCH: "arm64"},
	))
	if r.Status != Pass {
		t.Fatalf("expected Pass, got %v (%q)", r.Status, r.Message)
	}
	if r.Message != "All 2 release targets compile." {
		t.Fatalf("unexpected pass message: %q", r.Message)
	}
}

func TestCheckBuild_FailAggregatesSortedDetails(t *testing.T) {
	p := filepath.Join(t.TempDir(), "fail.sh")
	if err := os.WriteFile(p, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}
	r := checkBuild(buildCfg(p,
		config.BuildTarget{GOOS: "windows", GOARCH: "arm64"},
		config.BuildTarget{GOOS: "linux", GOARCH: "amd64"},
	))
	if r.Status != Fail {
		t.Fatalf("expected Fail, got %v", r.Status)
	}
	if r.Message != "These release targets failed to build:" {
		t.Fatalf("unexpected fail message: %q", r.Message)
	}
	if len(r.Details) != 2 {
		t.Fatalf("expected 2 details, got %d", len(r.Details))
	}
	if !strings.HasPrefix(r.Details[0], "linux/amd64") {
		t.Fatalf("details not sorted, first=%q", r.Details[0])
	}
	if !strings.HasPrefix(r.Details[1], "windows/arm64") {
		t.Fatalf("details not sorted, second=%q", r.Details[1])
	}
}
