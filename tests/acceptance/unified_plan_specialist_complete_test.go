// Acceptance: specs/unified-plan-specialist.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: Planner evidence with both artifacts passes plan complete
func TestUPS_PlannerEvidencePassesPlanComplete(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)
	upsWriteWorkflow(t, dir, "new-widget", "planner-v1")
	upsFeatureDocs(t, dir, "new-widget")
	upsWriteEvidence(t, dir, "new-widget", "planner", "senior-engineer",
		[]string{"Existing users keep working without migration"})

	out, code := runCent(t, bin, dir, "complete", "new-widget")
	if code != 0 {
		t.Fatalf("planner evidence must advance plan, got %d: %s", code, out)
	}
	if !strings.Contains(out, "code") {
		t.Fatalf("code step should become active: %s", out)
	}
}

// Scenario: A fresh workflow refuses a hand-authored legacy-role evidence pair
func TestUPS_ForgedLegacyPairBlockedNamingPlanner(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)
	upsWriteWorkflow(t, dir, "new-widget", "planner-v1")
	upsFeatureDocs(t, dir, "new-widget")
	// Hand-authored legacy pair, no planner evidence at all.
	upsWriteEvidence(t, dir, "new-widget", "big-thinker", "feature-specialist", []string{"n/a"})
	upsWriteEvidence(t, dir, "new-widget", "feature-specialist", "senior-engineer", []string{"n/a"})

	out, code := runCent(t, bin, dir, "complete", "new-widget")
	if code == 0 {
		t.Fatalf("a forged legacy pair must not satisfy a planner-v1 workflow: %s", out)
	}
	if !strings.Contains(out, "planner") {
		t.Fatalf("block message must name planner as the required role: %s", out)
	}
	if strings.Contains(out, "big-thinker") || strings.Contains(out, "feature-specialist") {
		t.Fatalf("a pinned workflow's gate must never reference the legacy pair: %s", out)
	}
}
