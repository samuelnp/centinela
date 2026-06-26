package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/teamdashboard"
)

// Blank Profile/Archetype/Worktree fall back to default/canonical/— at render.
func TestRenderDashboard_BlankFieldDefaults(t *testing.T) {
	d := teamdashboard.Dashboard{
		Features: []teamdashboard.FeatureRow{{
			Feature: "beta", Step: "tests", StepIndex: 2, StepTotal: 5, Owner: "unknown",
		}},
	}
	out := RenderDashboard(d)
	for _, want := range []string{"default", "canonical", "—", "owner unknown"} {
		if !strings.Contains(out, want) {
			t.Fatalf("blank-field default missing %q:\n%s", want, out)
		}
	}
}

// A present-but-empty roadmap renders the 0/0 done line, not an empty state.
func TestRenderDashboard_PresentEmptyRoadmap(t *testing.T) {
	d := teamdashboard.Dashboard{
		Roadmap: teamdashboard.RoadmapBurndown{Present: true, Done: 0, Total: 0},
	}
	out := RenderDashboard(d)
	if !strings.Contains(out, "0/0 done") {
		t.Fatalf("present-empty roadmap should render 0/0 done:\n%s", out)
	}
	if strings.Contains(out, "no roadmap") {
		t.Fatalf("present roadmap must not show empty state:\n%s", out)
	}
}
