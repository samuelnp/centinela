package acceptance_test

// Acceptance: specs/roadmap-phase-ops.feature
// Scenario: phase add with --after inserts immediately after the named phase
// Scenario: phase add without --after lands before the Backlog phase
// Scenario: phase add --note sets the phase note
// Scenario: phase add on an empty roadmap succeeds as the first phase
// Scenario: phase rename renames in place, leaving its features and other phases untouched
// Scenario: phase remove deletes an empty phase

import (
	"os"
	"strings"
	"testing"
)

const poAccBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"auth-service"}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"billing-api"}]},` +
	`{"name":"Backlog","features":[]}]}`

// poRoad returns the on-disk roadmap.json text for project d.
func poRoad(t *testing.T, d string) string {
	t.Helper()
	b, _ := os.ReadFile(emPath(d))
	return string(b)
}

// TestAcc_PhaseAddAfterAndNote drives add --after / --note through the binary.
func TestAcc_PhaseAddAfterAndNote(t *testing.T) {
	d := rmcProject(t, poAccBody)
	if _, _, code := rmcRun(t, d, "roadmap", "phase", "add", "Phase 1.5: Bridge",
		"--after", "Phase 1: Foundations", "--note", "bridge work"); code != 0 {
		t.Fatalf("phase add exit=%d", code)
	}
	road := poRoad(t, d)
	i1 := strings.Index(road, "Phase 1: Foundations")
	ib := strings.Index(road, "Phase 1.5: Bridge")
	i2 := strings.Index(road, "Phase 2: Growth")
	if !(i1 < ib && ib < i2) {
		t.Fatalf("bridge must sit between Phase 1 and Phase 2: %s", road)
	}
	if !strings.Contains(road, "bridge work") {
		t.Fatalf("note missing: %s", road)
	}
}

// TestAcc_PhaseAddDefaultBeforeBacklog: no --after lands before Backlog.
func TestAcc_PhaseAddDefaultBeforeBacklog(t *testing.T) {
	d := rmcProject(t, poAccBody)
	if _, _, code := rmcRun(t, d, "roadmap", "phase", "add", "Phase 3: Scale"); code != 0 {
		t.Fatalf("phase add exit=%d", code)
	}
	road := poRoad(t, d)
	if strings.Index(road, "Phase 3: Scale") > strings.Index(road, "Backlog") {
		t.Fatalf("Phase 3 must precede Backlog: %s", road)
	}
}

// TestAcc_PhaseAddEmptyRoadmap: add on {"phases":[]} becomes the first phase.
func TestAcc_PhaseAddEmptyRoadmap(t *testing.T) {
	d := rmcProject(t, `{"phases":[]}`)
	if _, _, code := rmcRun(t, d, "roadmap", "phase", "add", "Phase 1: Foundations"); code != 0 {
		t.Fatalf("phase add exit=%d", code)
	}
	if !strings.Contains(poRoad(t, d), "Phase 1: Foundations") {
		t.Fatal("first phase must be added")
	}
}

// TestAcc_PhaseRenameAndRemove: rename in place, then remove an empty phase.
func TestAcc_PhaseRenameAndRemove(t *testing.T) {
	d := rmcProject(t, poAccBody)
	if _, _, code := rmcRun(t, d, "roadmap", "phase", "rename", "Phase 1: Foundations", "Phase 1: Core"); code != 0 {
		t.Fatalf("rename exit=%d", code)
	}
	if road := poRoad(t, d); !strings.Contains(road, "Phase 1: Core") || !strings.Contains(road, "auth-service") {
		t.Fatalf("rename must keep features: %s", road)
	}
	if _, _, code := rmcRun(t, d, "roadmap", "phase", "add", "Phase 9: Empty"); code != 0 {
		t.Fatalf("add empty phase exit=%d", code)
	}
	if _, _, code := rmcRun(t, d, "roadmap", "phase", "remove", "Phase 9: Empty"); code != 0 {
		t.Fatalf("remove empty exit=%d", code)
	}
	if strings.Contains(poRoad(t, d), "Phase 9: Empty") {
		t.Fatal("empty phase must be removed")
	}
}
