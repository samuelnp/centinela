// Acceptance: specs/unified-plan-specialist.feature
package acceptance_test

import (
	"os/exec"
	"strings"
	"testing"
)

func upsStatusline(t *testing.T, bin, dir string) string {
	t.Helper()
	cmd := exec.Command(bin, "hook", "statusline")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("{}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook statusline failed: %v\n%s", err, out)
	}
	return string(out)
}

// Scenario: The statusline names planner for plan-step delegation
//
// The statusline protocol has no per-role field; what "names planner" means at
// this surface is that a role sub-workflow accidentally saved under
// "<feature>-planner" is never mistaken for the primary tracked feature (the
// isRoleWorkflow guard senior-engineer added), so "new-widget" — the feature a
// planner is actually delegated for — stays the one shown.
func TestUPS_StatuslineExcludesPlannerRoleWorkflow(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)
	upsWriteWorkflow(t, dir, "new-widget", "planner-v1")
	upsWriteWorkflow(t, dir, "new-widget-planner", "planner-v1")

	out := upsStatusline(t, bin, dir)
	if !strings.Contains(out, "WF:new-widget") {
		t.Fatalf("statusline must show the delegating feature new-widget: %s", out)
	}
	if strings.Contains(out, "WF:new-widget-planner") {
		t.Fatalf("statusline must never show the role sub-workflow as primary: %s", out)
	}
}

// Scenario: The statusline still names the legacy pair for an unpinned workflow
func TestUPS_StatuslineExcludesLegacyRoleWorkflows(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)
	upsWriteWorkflow(t, dir, "already-in-flight", "")
	upsWriteWorkflow(t, dir, "already-in-flight-big-thinker", "")
	upsWriteWorkflow(t, dir, "already-in-flight-feature-specialist", "")

	out := upsStatusline(t, bin, dir)
	if !strings.Contains(out, "WF:already-in-flight") {
		t.Fatalf("statusline must show already-in-flight: %s", out)
	}
	for _, gone := range []string{"WF:already-in-flight-big-thinker", "WF:already-in-flight-feature-specialist"} {
		if strings.Contains(out, gone) {
			t.Fatalf("statusline must never show a legacy role sub-workflow as primary: %s", out)
		}
	}
}
