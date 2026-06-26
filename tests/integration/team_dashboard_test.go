package integration_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/teamdashboard"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

func realisticInputs() teamdashboard.Inputs {
	now := time.Date(2026, 6, 26, 0, 0, 0, 0, time.UTC)
	return teamdashboard.Inputs{
		Active: []*workflow.Workflow{{
			Feature: "alpha", CurrentStep: "code", StartedAt: now.Add(-72 * time.Hour),
			StepOrder: []string{"plan", "code", "tests", "validate", "docs"},
			EnforcementProfile: "strict", Archetype: "hexagonal", WorktreePath: ".worktrees/alpha",
		}},
		Roadmap: &roadmap.Roadmap{Phases: []roadmap.Phase{
			{Name: "Q1", Features: []roadmap.Feature{{Name: "f1"}, {Name: "f2"}}},
		}},
		Events: []telemetry.Event{
			{Type: telemetry.TypeGateFailure, Gate: "coverage"},
			{Type: telemetry.TypeGateFailure, Gate: "coverage"},
		},
		Owners: map[string]string{"alpha": "Alice"},
		Now:    now,
	}
}

// Compute round-trips through --json's MarshalIndent shape with stable keys.
func TestTeamDashboard_ComputeJSONRoundTrip(t *testing.T) {
	d := teamdashboard.Compute(realisticInputs())
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	for _, f := range []string{"Features", "Roadmap", "Gates"} {
		if _, ok := m[f]; !ok {
			t.Fatalf("missing top-level field %q: %v", f, m)
		}
	}
	var back teamdashboard.Dashboard
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if len(back.Features) != 1 || back.Features[0].Owner != "Alice" {
		t.Fatalf("round-trip features: %+v", back.Features)
	}
	if back.Roadmap.Total != 2 || len(back.Gates) != 1 {
		t.Fatalf("round-trip roadmap/gates: %+v %+v", back.Roadmap, back.Gates)
	}
}

// RenderDashboard over the same Inputs is stable and shows all three panels.
func TestTeamDashboard_RenderStability(t *testing.T) {
	d := teamdashboard.Compute(realisticInputs())
	a := ui.RenderDashboard(d)
	b := ui.RenderDashboard(d)
	if a != b {
		t.Fatalf("render not stable:\n%s\n---\n%s", a, b)
	}
	for _, want := range []string{"In-flight features", "Roadmap burn-down", "Gate health", "alpha", "coverage"} {
		if !strings.Contains(a, want) {
			t.Fatalf("render missing %q:\n%s", want, a)
		}
	}
}
