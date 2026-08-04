package unit_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

// paddedLines returns every line of s that still ends in a space or a tab.
func paddedLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if line != strings.TrimRight(line, " \t") {
			out = append(out, line)
		}
	}
	return out
}

// dietWfs is a mismatched-width fixture: lipgloss.JoinVertical only pads when
// sibling lines differ in length, so equal-width features would make these
// assertions vacuous.
func dietWfs() []*workflow.Workflow {
	return []*workflow.Workflow{
		{Feature: "a", CurrentStep: "plan"},
		{Feature: "urgent-fix", CurrentStep: "code", Archetype: "hotfix",
			StepOrder: []string{"code", "tests", "validate"}},
		{Feature: "a-considerably-longer-feature-slug", CurrentStep: "docs",
			StepOrder: []string{"plan", "code", "tests", "validate", "docs"}},
	}
}

// TestHookOnlyPanelsAreUnpadded is the unit-tier statement of the feature's
// whole contract, exercised through the exported package surface only.
func TestHookOnlyPanelsAreUnpadded(t *testing.T) {
	renders := map[string]string{
		"RenderContextCapped":        ui.RenderContextCapped(dietWfs(), 2),
		"RenderReviewReady":          ui.RenderReviewReady("a-considerably-longer-feature-slug", "docs", "done"),
		"RenderFeatureBriefNeeded":   ui.RenderFeatureBriefNeeded("a"),
		"RenderEdgeCaseReportNeeded": ui.RenderEdgeCaseReportNeeded("urgent-fix"),
		"RenderDocumentationNeeded":  ui.RenderDocumentationNeeded("landing-page-content"),
		"RenderChangelogNeeded":      ui.RenderChangelogNeeded("urgent-fix"),
	}
	for name, out := range renders {
		if bad := paddedLines(out); len(bad) > 0 {
			t.Errorf("%s emitted %d padded line(s), first: %q", name, len(bad), bad[0])
		}
	}
}

// TestHookPanelKeepsArchetypeLadder pins the governance signal the plan refused
// to cut: the per-archetype step ladder, which is the only per-turn indication
// of whether a validate gate applies to a given workflow at all.
func TestHookPanelKeepsArchetypeLadder(t *testing.T) {
	out := ui.RenderContextCapped(dietWfs(), 2)
	for _, want := range []string{"a-considerably-longer-feature-slug", "urgent-fix", "+2 more active"} {
		if !strings.Contains(out, want) {
			t.Errorf("active-workflows panel lost %q:\n%s", want, out)
		}
	}
	spike := ui.RenderContextCapped([]*workflow.Workflow{
		{Feature: "quick-check", CurrentStep: "code", Archetype: "spike",
			StepOrder: []string{"plan", "code"}},
	}, 0)
	if strings.Contains(spike, "validate") {
		t.Errorf("spike ladder must not advertise a validate step it never gates on:\n%s", spike)
	}
	if !strings.Contains(spike, "code 1/2") {
		t.Errorf("spike progress lost:\n%s", spike)
	}
}

// TestCLIRenderStatusStillPadded pins the scope boundary at the tier that a
// future cleanup is most likely to touch by accident: the CLI surface keeps its
// bytes exactly, because this feature changed only what the hook emits.
func TestCLIRenderStatusStillPadded(t *testing.T) {
	out := ui.RenderStatus(&workflow.Workflow{Feature: "sample-feature", CurrentStep: "plan"})
	if len(paddedLines(out)) == 0 {
		t.Errorf("RenderStatus lost its padding — the hook panel diet was scoped to "+
			"hook-only renderers; changing centinela status output is a separate decision:\n%s", out)
	}
}
