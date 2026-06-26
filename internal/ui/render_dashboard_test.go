package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/teamdashboard"
)

func TestRenderDashboard_PopulatedPanels(t *testing.T) {
	d := teamdashboard.Dashboard{
		Features: []teamdashboard.FeatureRow{{
			Feature: "alpha", Step: "code", StepIndex: 1, StepTotal: 5,
			AgeDays: 3, Profile: "strict", Archetype: "hexagonal",
			Worktree: ".worktrees/alpha", Owner: "Alice",
		}},
		Roadmap: teamdashboard.RoadmapBurndown{
			Present: true, Done: 3, Total: 7,
			Phases: []teamdashboard.PhaseStatus{{Name: "Q1", Done: 2, Total: 3}},
		},
		Gates: []teamdashboard.GateHealth{{Gate: "coverage", Fails: 2}},
	}
	out := RenderDashboard(d)
	for _, want := range []string{
		"Team Dashboard", "In-flight features", "Roadmap burn-down", "Gate health",
		"alpha", "code 1/5", "3d", "strict", "hexagonal", ".worktrees/alpha", "Alice",
		"Q1", "2/3", "3/7 done", "coverage",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("populated output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("non-TTY output must be ANSI-free: %q", out)
	}
}

func TestRenderDashboard_AllEmptyStates(t *testing.T) {
	out := RenderDashboard(teamdashboard.Dashboard{})
	for _, want := range []string{
		"no active features", "no roadmap", "no gate failures recorded",
		"0 active · no roadmap",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("empty-state output missing %q:\n%s", want, out)
		}
	}
}
