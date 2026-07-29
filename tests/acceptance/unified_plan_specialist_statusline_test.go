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

// Scenario: A planner role sub-workflow never displaces the feature on the statusline
//
// The statusline has no per-step role field — adding one is deferred as
// `statusline-plan-role-display`. Both halves below are asserted exactly as the
// scenario states them: the surface tracks the delegating feature and never the
// "-planner" sub-workflow, and the hook directive is where the role is named.
func TestUPS_StatuslineExcludesPlannerRoleWorkflow(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)
	upsWriteWorkflow(t, dir, "new-widget", "planner-v1")
	upsWriteWorkflow(t, dir, "new-widget-planner", "planner-v1")

	out := upsStatusline(t, bin, dir)
	if !strings.Contains(out, "WF:new-widget") {
		t.Fatalf("statusline must track the delegating feature new-widget: %s", out)
	}
	if strings.Contains(out, "WF:new-widget-planner") {
		t.Fatalf("statusline must never promote the role sub-workflow to primary: %s", out)
	}

	// And the hook directive for "new-widget" should name the role "planner".
	upsFeatureDocs(t, dir, "new-widget")
	directive := upsDirective(t, bin, dir)
	if !strings.Contains(directive, "delegate to [planner") {
		t.Fatalf("directive must name planner: %s", directive)
	}
	for _, retired := range []string{"big-thinker", "feature-specialist"} {
		if strings.Contains(directive, retired) {
			t.Fatalf("pinned directive must not name %q: %s", retired, directive)
		}
	}
}

// Scenario: Legacy role sub-workflows never displace an unpinned feature
//
// The directive half is asserted as the scenario states it — BOTH legacy roles
// named, planner not offered. The previous revision of this test asserted the
// inverse of its scenario text while the traceability gate reported it covered.
func TestUPS_StatuslineExcludesLegacyRoleWorkflows(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)
	upsWriteWorkflow(t, dir, "already-in-flight", "")
	upsWriteWorkflow(t, dir, "already-in-flight-big-thinker", "")
	upsWriteWorkflow(t, dir, "already-in-flight-feature-specialist", "")

	out := upsStatusline(t, bin, dir)
	if !strings.Contains(out, "WF:already-in-flight") {
		t.Fatalf("statusline must track already-in-flight: %s", out)
	}
	for _, gone := range []string{"WF:already-in-flight-big-thinker", "WF:already-in-flight-feature-specialist"} {
		if strings.Contains(out, gone) {
			t.Fatalf("statusline must never promote a legacy role sub-workflow to primary: %s", out)
		}
	}

	upsFeatureDocs(t, dir, "already-in-flight")
	directive := upsDirective(t, bin, dir)
	for _, want := range []string{"big-thinker", "feature-specialist"} {
		if !strings.Contains(directive, want) {
			t.Fatalf("unpinned directive must name %q: %s", want, directive)
		}
	}
	if strings.Contains(directive, "delegate to [planner") ||
		strings.Contains(directive, "already-in-flight-planner") {
		t.Fatalf("unpinned directive must not offer planner: %s", directive)
	}
}
