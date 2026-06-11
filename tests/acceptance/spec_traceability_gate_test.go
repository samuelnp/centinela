// Acceptance: specs/spec-traceability-gate.feature
package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gates"
)

const stgFeature = "Feature: f\n  Scenario: Start the watcher\n"

// Scenario: A scenario with a matching acceptance test passes the gate
func TestSTG_MatchingTestPasses(t *testing.T) {
	t.Chdir(t.TempDir())
	stgWrite(t, "specs/x.feature", stgFeature)
	stgWrite(t, "tests/acceptance/x_test.go", "// Acceptance: specs/x.feature\n// Scenario: Start the watcher\n")
	r := stgRun(t, stgConfig("fail"), nil)
	if r.Status != gates.Pass || !strings.Contains(r.Message, "1 scenarios") {
		t.Fatalf("want Pass with count, got %v %q", r.Status, r.Message)
	}
}

// Scenario: A scenario with no acceptance test fails the gate
func TestSTG_UncoveredFails(t *testing.T) {
	t.Chdir(t.TempDir())
	stgWrite(t, "specs/x.feature", stgFeature)
	r := stgRun(t, stgConfig("fail"), nil)
	if r.Status != gates.Fail || !strings.Contains(detailsJoined(r), `specs/x.feature: "Start the watcher"`) {
		t.Fatalf("want Fail naming spec+scenario, got %v %v", r.Status, r.Details)
	}
}

// Scenario: Matching normalizes trailing period, spacing, and letter case
func TestSTG_NormalizationMatches(t *testing.T) {
	t.Chdir(t.TempDir())
	stgWrite(t, "specs/x.feature", stgFeature)
	stgWrite(t, "tests/acceptance/x_test.go", "// Acceptance: specs/x.feature\n// Scenario:  start the WATCHER .\n")
	if r := stgRun(t, stgConfig("fail"), nil); r.Status != gates.Pass {
		t.Fatalf("normalized comment must match, got %v %v", r.Status, r.Details)
	}
}

// Scenario: An acceptance header with a trailing annotation still matches its spec
func TestSTG_HeaderAnnotationMatches(t *testing.T) {
	t.Chdir(t.TempDir())
	stgWrite(t, "specs/x.feature", stgFeature)
	stgWrite(t, "tests/acceptance/x_test.go", "// Acceptance: specs/x.feature (AC4, AC5)\n// Scenario: Start the watcher\n")
	if r := stgRun(t, stgConfig("fail"), nil); r.Status != gates.Pass {
		t.Fatalf("trailing annotation must be ignored, got %v %v", r.Status, r.Details)
	}
}

// Scenario: A Scenario Outline counts as one covered scenario
func TestSTG_ScenarioOutlineCountsOnce(t *testing.T) {
	t.Chdir(t.TempDir())
	stgWrite(t, "specs/x.feature", "Feature: f\n  Scenario Outline: Run rows\n    Examples:\n      | a |\n      | 1 |\n")
	stgWrite(t, "tests/acceptance/x_test.go", "// Acceptance: specs/x.feature\n// Scenario: Run rows\n")
	r := stgRun(t, stgConfig("fail"), nil)
	if r.Status != gates.Pass || !strings.Contains(r.Message, "1 scenarios") {
		t.Fatalf("outline must count once and pass, got %v %q", r.Status, r.Message)
	}
}
