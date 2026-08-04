// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"path/filepath"
	"testing"
)

// Scenario: no nudge when the Backlog is empty
func TestRsh_NoNudgeWhenBacklogEmpty(t *testing.T) {
	bin := buildCent(t)
	dir := rshNudgeRepo(t, []string{"last-one"}, 0)
	rshDocsWorkflow(t, dir, "last-one")

	out, code := runCent(t, bin, dir, "complete", "last-one")
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	if contains(out, "deferred findings remain") {
		t.Fatalf("nudge must not print for an empty Backlog:\n%s", out)
	}
}

// Scenario: an unreadable roadmap suppresses the nudge without failing complete
func TestRsh_UnreadableRoadmapSuppressesNudge(t *testing.T) {
	bin := buildCent(t)
	dir := rshNudgeRepo(t, []string{"last-one"}, 3)
	rshDocsWorkflow(t, dir, "last-one")
	mustWrite(t, filepath.Join(dir, ".workflow", "roadmap.json"), "{not json")

	out, code := runCent(t, bin, dir, "complete", "last-one")
	if code != 0 {
		t.Fatalf("an unreadable roadmap must not fail complete: exit=%d\n%s", code, out)
	}
	if contains(out, "deferred findings remain") {
		t.Fatalf("nudge must be suppressed, not printed:\n%s", out)
	}
}
