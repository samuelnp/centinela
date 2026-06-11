// Acceptance: specs/spec-traceability-gate.feature
package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// Scenario: Warn severity reports gaps without failing
func TestSTG_WarnSeverityDoesNotFail(t *testing.T) {
	t.Chdir(t.TempDir())
	stgWrite(t, "specs/x.feature", stgFeature)
	r := stgRun(t, stgConfig("warn"), nil)
	if r.Status != gates.Warn || !strings.Contains(detailsJoined(r), "Start the watcher") {
		t.Fatalf("want Warn listing gap, got %v %v", r.Status, r.Details)
	}
}

// Scenario: Diff-aware scope gates only changed spec files
func TestSTG_DiffAwareScopesChangedSpecs(t *testing.T) {
	t.Chdir(t.TempDir())
	stgWrite(t, "specs/unchanged.feature", "Feature: u\n  Scenario: Never covered\n")
	stgWrite(t, "specs/changed.feature", "Feature: c\n  Scenario: Covered here\n")
	stgWrite(t, "tests/acceptance/x_test.go", "// Acceptance: specs/changed.feature\n// Scenario: Covered here\n")
	filter := gitdiff.NewSet([]string{filepath.Join("specs", "changed.feature")})
	if r := stgRun(t, stgConfig("fail"), filter); r.Status != gates.Pass {
		t.Fatalf("unchanged spec must be out of scope, got %v %v", r.Status, r.Details)
	}
}

// Scenario: No spec files in scope skips the gate
func TestSTG_NoSpecsSkips(t *testing.T) {
	t.Chdir(t.TempDir())
	if r := stgRun(t, stgConfig("fail"), nil); r.Status != gates.Skip {
		t.Fatalf("want Skip when no specs, got %v %q", r.Status, r.Message)
	}
}

// Scenario: An unknown severity value is rejected at config load
func TestSTG_UnknownSeverityRejected(t *testing.T) {
	t.Chdir(t.TempDir())
	stgWrite(t, config.Filename, "[gates.spec_traceability]\nenabled=true\nseverity=\"loud\"\n")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "severity") {
		t.Fatalf("want severity error, got %v", err)
	}
}

// Scenario: The gate is registered and enabled for Centinela in warn mode
func TestSTG_CentinelaEnablesWarn(t *testing.T) {
	t.Chdir(filepath.Join("..", "..")) // repo root, where centinela.toml lives
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	g := cfg.Gates.SpecTraceability
	if !g.Enabled || g.Severity != "warn" {
		t.Fatalf("repo toml must enable gate in warn mode, got %+v", g)
	}
}
