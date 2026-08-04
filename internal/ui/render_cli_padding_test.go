package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/workflow"
)

// paddedLineCount reports how many lines still carry lipgloss right-padding.
func paddedLineCount(s string) int {
	n := 0
	for _, line := range strings.Split(s, "\n") {
		if line != strings.TrimRight(line, " \t") {
			n++
		}
	}
	return n
}

// TestCLIRenderersKeepTheirPadding pins the scope boundary of the hook panel
// diet: it changed what the HOOK emits, not how the CLI renders to a TTY.
// RenderStatus and RenderBlocked still compose via lipgloss.JoinVertical and
// must keep every padded byte they had before. This also catches the most
// likely accidental over-reach — moving the trim down into the shared
// renderSystemPanel/JoinVertical primitives, which would silently change
// `centinela status` and blocked-write output for every user.
//
// Cleaning these up is defensible, but it is a separate, deliberate decision
// (backlog: hook-panel-diet-remaining-emitters). If you are here because this
// test went red, that is the conversation to have — not a fix to apply.
func TestCLIRenderersKeepTheirPadding(t *testing.T) {
	wf := &workflow.Workflow{
		Feature: "sample-feature", CurrentStep: "plan", StartedAt: time.Unix(0, 0).UTC(),
	}
	cases := map[string]string{
		"RenderStatus":  RenderStatus(wf),
		"RenderBlocked": RenderBlocked("code", "plan", "sample-feature", "/tmp/a.go"),
	}
	for name, out := range cases {
		if paddedLineCount(out) == 0 {
			t.Errorf("%s lost its trailing-whitespace padding — this feature scoped the "+
				"trim to hook-only render functions. Was trimTrailingWS moved into "+
				"renderSystemPanel or into lipgloss composition?\n%s", name, out)
		}
	}
}

// TestRenderRoadmapUnaffected records the honest finding that RenderRoadmap has
// no padding to preserve: it composes with strings.Join, never
// lipgloss.JoinVertical, so it emitted zero trailing whitespace before this
// feature and emits zero after. Asserting "still padded" for it would be false.
// The shared-primitive regression it would otherwise guard is already covered
// by TestCLIRenderersKeepTheirPadding above.
func TestRenderRoadmapUnaffected(t *testing.T) {
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "P1", Features: []roadmap.Feature{
		{Name: "widget-a"}, {Name: "a-considerably-longer-feature-name"},
	}}}}
	out := RenderRoadmap(r)
	if paddedLineCount(out) != 0 {
		t.Errorf("RenderRoadmap unexpectedly padded — it composes with strings.Join; "+
			"if that changed, re-read this test's premise:\n%s", out)
	}
	for _, want := range []string{"PHASE OVERVIEW", "P1", "widget-a"} {
		if !strings.Contains(out, want) {
			t.Errorf("RenderRoadmap lost %q:\n%s", want, out)
		}
	}
}
