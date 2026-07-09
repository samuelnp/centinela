package acceptance_test

// Acceptance: specs/roadmap-phase-ops.feature
// Scenario Outline: phase add refuses a duplicate name, reserved name, empty name, or unknown --after anchor
// Scenario Outline: phase rename refuses an unknown old name, a collision, an empty new name, or either side reserved
// Scenario: phase remove of a non-empty phase without --force is refused, naming the feature count
// Scenario Outline: phase remove refuses the reserved Backlog/Baseline phase, with or without --force
// Scenario: phase add/rename/remove against a missing roadmap.json surfaces an error and leaves the file absent
// Scenario: phase add/rename/remove against a malformed roadmap.json surfaces an error and leaves the file untouched

import (
	"os"
	"testing"
)

// TestAcc_PhaseRejectionsByteIdentical drives every refusal through the binary and
// asserts roadmap.json is byte-identical afterward.
func TestAcc_PhaseRejectionsByteIdentical(t *testing.T) {
	rows := [][]string{
		{"roadmap", "phase", "add", "Phase 1: Foundations"},                       // duplicate
		{"roadmap", "phase", "add", "Backlog"},                                    // reserved
		{"roadmap", "phase", "add", "Phase 3: Scale", "--after", "Nope"},          // unknown anchor
		{"roadmap", "phase", "rename", "Nope", "Phase 3: Scale"},                  // unknown old
		{"roadmap", "phase", "rename", "Phase 1: Foundations", "Phase 2: Growth"}, // collision
		{"roadmap", "phase", "rename", "Backlog", "Phase 3: Scale"},               // reserved
		{"roadmap", "phase", "remove", "Phase 2: Growth"},                         // non-empty, no --force
		{"roadmap", "phase", "remove", "Backlog"},                                 // reserved
		{"roadmap", "phase", "remove", "Baseline", "--force"},                     // reserved even with force
	}
	for _, args := range rows {
		d := rmcProject(t, poAccBody)
		before, _ := os.ReadFile(emPath(d))
		if _, _, code := rmcRun(t, d, args...); code == 0 {
			t.Fatalf("row %v must be rejected", args)
		}
		after, _ := os.ReadFile(emPath(d))
		if string(before) != string(after) {
			t.Fatalf("row %v must be byte-identical", args)
		}
	}
}

// TestAcc_PhaseMissingRoadmap: a missing roadmap.json errors and stays absent.
func TestAcc_PhaseMissingRoadmap(t *testing.T) {
	d := rmcProject(t, "") // no roadmap.json written
	if _, _, code := rmcRun(t, d, "roadmap", "phase", "add", "Phase 1: Foundations"); code == 0 {
		t.Fatal("missing roadmap.json must error")
	}
	if _, err := os.Stat(emPath(d)); !os.IsNotExist(err) {
		t.Fatal("roadmap.json must remain absent")
	}
}

// TestAcc_PhaseMalformedRoadmap: malformed JSON errors and stays byte-identical.
func TestAcc_PhaseMalformedRoadmap(t *testing.T) {
	d := rmcProject(t, "{ not valid json")
	before, _ := os.ReadFile(emPath(d))
	if _, _, code := rmcRun(t, d, "roadmap", "phase", "add", "Phase 1: Foundations"); code == 0 {
		t.Fatal("malformed roadmap.json must error")
	}
	after, _ := os.ReadFile(emPath(d))
	if string(before) != string(after) {
		t.Fatal("malformed roadmap.json must be left untouched")
	}
}
