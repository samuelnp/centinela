package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// Acceptance spec: specs/spec-traceability-gate.feature
//
// Integration coverage runs the gate end-to-end through the public
// gates.RunWithFilter surface against a throwaway specs+acceptance tree on disk.

func stgIntWrite(t *testing.T, rel, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(rel), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rel, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func stgIntCfg() *config.Config {
	cfg := &config.Config{}
	cfg.Gates.SpecTraceability = config.SpecTraceabilityConfig{
		Enabled: true, SpecDir: "specs", TestDir: "tests/acceptance", Severity: "fail",
	}
	return cfg
}

func stgIntResult(t *testing.T, filter *gitdiff.Set) gates.Result {
	t.Helper()
	for _, r := range gates.RunWithFilter(stgIntCfg(), filter) {
		if r.Name == "spec-traceability-gate" {
			return r
		}
	}
	t.Fatal("spec-traceability-gate result missing")
	return gates.Result{}
}

func TestIntegration_SpecTraceability_CoveredPasses(t *testing.T) {
	t.Chdir(t.TempDir())
	stgIntWrite(t, filepath.Join("specs", "w.feature"), "Feature: f\n  Scenario: Watch\n")
	stgIntWrite(t, filepath.Join("tests", "acceptance", "w_test.go"),
		"// Acceptance: specs/w.feature\n// Scenario: Watch\n")
	if r := stgIntResult(t, nil); r.Status != gates.Pass {
		t.Fatalf("covered tree must Pass end-to-end, got %v %v", r.Status, r.Details)
	}
}

func TestIntegration_SpecTraceability_UncoveredFails(t *testing.T) {
	t.Chdir(t.TempDir())
	stgIntWrite(t, filepath.Join("specs", "w.feature"), "Feature: f\n  Scenario: Watch\n")
	r := stgIntResult(t, nil)
	if r.Status != gates.Fail || !strings.Contains(strings.Join(r.Details, "\n"), "Watch") {
		t.Fatalf("uncovered tree must Fail naming gap, got %v %v", r.Status, r.Details)
	}
}
