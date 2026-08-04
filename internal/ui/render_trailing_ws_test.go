package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/samuelnp/centinela/internal/workflow"
)

// dietWorkflows is the realistic multi-workflow fixture: five active features
// (the hook's display cap), deliberately mismatched in name length and step
// order, because lipgloss.JoinVertical only pads when sibling lines differ in
// width — a same-width fixture would pass these assertions even with the fix
// reverted. Archetypes differ so the step ladders differ in width too.
func dietWorkflows() []*workflow.Workflow {
	return []*workflow.Workflow{
		{Feature: "a", CurrentStep: "plan"},
		{Feature: "hook-context-panel-diet", CurrentStep: "tests",
			StepOrder: []string{"plan", "code", "tests", "validate", "docs"},
			Steps: map[string]workflow.StepState{
				"plan": {Status: "done"}, "code": {Status: "done"}}},
		{Feature: "urgent-fix", CurrentStep: "code", Archetype: "hotfix",
			StepOrder: []string{"code", "tests", "validate"}},
		{Feature: "quick-check", CurrentStep: "code", Archetype: "spike",
			StepOrder: []string{"plan", "code"}},
		{Feature: "a-considerably-longer-feature-slug-than-the-rest", CurrentStep: "docs",
			StepOrder: []string{"plan", "code", "tests", "validate", "docs"}},
	}
}

// dietRenders returns every hook-only panel this feature trims, under fixtures
// that cover the multi-workflow nudge loop, an empty list, a done workflow and
// a user-facing feature.
func dietRenders() map[string]string {
	wfs := dietWorkflows()
	done := &workflow.Workflow{Feature: "shipped-feature", CurrentStep: "done"}
	return map[string]string{
		"RenderContextCapped/five active + capped remainder": RenderContextCapped(wfs, 2),
		"RenderContextCapped/single workflow":                RenderContextCapped(wfs[:1], 0),
		"RenderContextCapped/no active workflows":            RenderContextCapped(nil, 0),
		"RenderContextCapped/done workflow":                  RenderContextCapped([]*workflow.Workflow{done}, 0),
		"RenderReviewReady/long feature, short next":         RenderReviewReady(wfs[4].Feature, "docs", "done"),
		"RenderReviewReady/short feature":                    RenderReviewReady("a", "plan", "code"),
		"RenderFeatureBriefNeeded/long feature":              RenderFeatureBriefNeeded(wfs[4].Feature),
		"RenderFeatureBriefNeeded/short feature":             RenderFeatureBriefNeeded("a"),
		"RenderEdgeCaseReportNeeded":                         RenderEdgeCaseReportNeeded(wfs[1].Feature),
		"RenderDocumentationNeeded/user-facing":              RenderDocumentationNeeded("landing-page-content"),
		"RenderChangelogNeeded/internal":                     RenderChangelogNeeded(wfs[1].Feature),
	}
}

func TestHookPanelsCarryNoTrailingWhitespace(t *testing.T) {
	for name, out := range dietRenders() {
		t.Run(name, func(t *testing.T) {
			for i, line := range strings.Split(out, "\n") {
				if line != strings.TrimRight(line, " \t") {
					t.Fatalf("line %d ends in whitespace: %q\nfull render:\n%s", i+1, line, out)
				}
			}
		})
	}
}

// TestHookPanelsKeepGovernanceSignal pins that trimming removed padding only:
// the feature name, step, progress and full step ladder must all survive.
func TestHookPanelsKeepGovernanceSignal(t *testing.T) {
	out := RenderContextCapped(dietWorkflows(), 2)
	for _, want := range []string{
		"hook-context-panel-diet", "tests 2/5", "quick-check", "urgent-fix",
		"+2 more active", "plan", "code", "validate", "docs",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("active-workflows panel lost %q:\n%s", want, out)
		}
	}
}

// TestDietFixturesActuallyPadWhenUntrimmed is the non-tautology check. It
// composes one fixture's panel body exactly as RenderContextCapped does but
// without the trim, and asserts that raw form IS padded — so the assertions
// above have real padding to remove rather than passing by fixture accident.
func TestDietFixturesActuallyPadWhenUntrimmed(t *testing.T) {
	wf := dietWorkflows()[4]
	raw := lipgloss.JoinVertical(lipgloss.Left, StyleBold.Render(wf.Feature), stepBar(wf))
	padded := false
	for _, line := range strings.Split(raw, "\n") {
		if line != strings.TrimRight(line, " \t") {
			padded = true
		}
	}
	if !padded {
		t.Fatal("fixture no longer exercises JoinVertical padding: the zero-trailing-whitespace " +
			"assertions in this file would now pass even with trimTrailingWS reverted to identity. " +
			"Widen the sibling line lengths in dietWorkflows().")
	}
}
