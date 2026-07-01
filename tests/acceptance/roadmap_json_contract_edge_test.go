package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/roadmap-json-contract.feature
// Scenario: roadmap show --json emits the persisted Roadmap verbatim, including non-schedulable phases
// Scenario: roadmap list --json is an alias for roadmap show --json
// Scenario: roadmap show (no flag) prints the same text as roadmap (no flag)
// Scenario: roadmap show --json is byte-identical across two consecutive runs
func TestRoadmapShowJSONVerbatimAndAlias(t *testing.T) {
	d := rmcProject(t, rmcBody)
	show, _, code := rmcRun(t, d, "roadmap", "show", "--json")
	if code != 0 {
		t.Fatalf("show --json exit=%d", code)
	}
	if !strings.Contains(show, `"Backlog"`) {
		t.Fatalf("show --json must retain non-schedulable Backlog:\n%s", show)
	}
	if strings.Contains(show, `"status"`) || strings.Contains(show, `"readiness"`) {
		t.Fatalf("show --json must carry no derived fields:\n%s", show)
	}
	if list, _, _ := rmcRun(t, d, "roadmap", "list", "--json"); list != show {
		t.Fatalf("list --json must be byte-identical to show --json")
	}
	if show2, _, _ := rmcRun(t, d, "roadmap", "show", "--json"); show2 != show {
		t.Fatalf("show --json must be byte-identical across runs")
	}
	showText, _, _ := rmcRun(t, d, "roadmap", "show")
	plainText, _, _ := rmcRun(t, d, "roadmap")
	if showText != plainText {
		t.Fatalf("roadmap show text must equal roadmap text")
	}
}

// Acceptance: specs/roadmap-json-contract.feature
// Scenario: Phase with zero features renders as an empty features array
// Scenario: Empty roadmap --json emits empty phases and all-zero counts
// Scenario: Empty roadmap ready --json emits an empty array
// Scenario: Empty roadmap show --json emits the persisted empty structure verbatim
func TestRoadmapEmptyAndZeroFeatureContracts(t *testing.T) {
	d := rmcProject(t, `{"phases":[]}`)
	out, _, code := rmcRun(t, d, "roadmap", "--json")
	if code != 0 {
		t.Fatalf("empty roadmap --json exit=%d", code)
	}
	if strings.Join(strings.Fields(out), "") != `{"phases":[],"counts":{"planned":0,"inProgress":0,"done":0}}` {
		t.Fatalf("empty view = %s", out)
	}
	ready, _, _ := rmcRun(t, d, "roadmap", "ready", "--json")
	if strings.TrimSpace(ready) != "[]" {
		t.Fatalf("empty ready must be [] not null, got %q", ready)
	}
	if _, _, c := rmcRun(t, d, "roadmap", "show", "--json"); c != 0 {
		t.Fatalf("empty show --json must exit 0, got %d", c)
	}
	zero := rmcProject(t, `{"phases":[{"name":"Q2","features":[]}]}`)
	zout, _, _ := rmcRun(t, zero, "roadmap", "--json")
	if !strings.Contains(strings.Join(strings.Fields(zout), ""), `"features":[]`) {
		t.Fatalf("zero-feature phase must render empty features array:\n%s", zout)
	}
}
