package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/setup"
)

// TestDateOnlyShortTimestamp covers the dateOnly fallback when the timestamp is
// shorter than a full date prefix, via spanLabel.
func TestDateOnlyShortTimestamp(t *testing.T) {
	if got := spanLabel("2026", "2026"); !strings.Contains(got, "2026 through 2026") {
		t.Fatalf("expected short timestamps passed through verbatim: %q", got)
	}
	if got := dateOnly("2026-06-30T12:00:00Z"); got != "2026-06-30" {
		t.Fatalf("expected truncated date, got %q", got)
	}
}

// TestRenderRoadmapJSONNeeded_InvalidBranch covers the non-IsNotExist arm where
// the file exists but is invalid.
func TestRenderRoadmapJSONNeeded_InvalidBranch(t *testing.T) {
	out := RenderRoadmapJSONNeeded(errors.New("unexpected end of JSON input"))
	if !strings.Contains(out, "invalid") {
		t.Fatalf("expected invalid-roadmap guidance: %q", out)
	}
}

// TestRenderRoadmapBacklogSection covers the IsBacklogPhaseName skip in the phase
// loop plus the non-empty renderBacklogSection block.
func TestRenderRoadmapBacklogSection(t *testing.T) {
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{
		{Name: "Phase 1", Features: []roadmap.Feature{{Name: "alpha"}}},
		{Name: roadmap.BacklogPhaseName, Features: []roadmap.Feature{
			{Name: "deferred-fix", Summary: "found during audit"},
		}},
	}}
	out := RenderRoadmap(r)
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "Phase 1") {
		t.Fatalf("expected schedulable phase rendered: %q", out)
	}
	if !strings.Contains(out, "deferred-fix") || !strings.Contains(out, "found during audit") {
		t.Fatalf("expected backlog section rendered: %q", out)
	}
}

// TestRenderSetupMigrationPlanApplyMode covers the applying=true branch.
func TestRenderSetupMigrationPlanApplyMode(t *testing.T) {
	plan := setup.SyncPlan{Items: []setup.SyncItem{{Path: "AGENTS.md", Action: setup.SyncCreate}}}
	if out := RenderSetupMigrationPlan(plan, true); !strings.Contains(out, "SETUP APPLY") {
		t.Fatalf("expected apply mode header: %q", out)
	}
}

// TestPanelStyleErrorTone covers the toneError branch of panelStyle.
func TestPanelStyleErrorTone(t *testing.T) {
	if panelStyle(toneError).Render("x") == "" {
		t.Fatal("error-tone panel style should render")
	}
}
