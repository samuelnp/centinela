// Acceptance: specs/unified-plan-specialist.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: A legacy in-flight workflow still requires and accepts the complete two-role set
func TestUPS_LegacyCompletePairAdvancesPlan(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)
	upsWriteWorkflow(t, dir, "already-in-flight", "") // empty PlanContract == legacy
	upsFeatureDocs(t, dir, "already-in-flight")
	upsWriteEvidence(t, dir, "already-in-flight", "big-thinker", "feature-specialist", []string{"n/a"})
	upsWriteEvidence(t, dir, "already-in-flight", "feature-specialist", "senior-engineer", []string{"n/a"})

	out, code := runCent(t, bin, dir, "complete", "already-in-flight")
	if code != 0 {
		t.Fatalf("complete legacy pair must advance plan, got %d: %s", code, out)
	}
}

// Scenario: A partial legacy set still fails, naming this workflow's required set
func TestUPS_LegacyPartialPairFailsWithContractAnnotation(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)
	upsWriteWorkflow(t, dir, "already-in-flight", "")
	upsFeatureDocs(t, dir, "already-in-flight")
	upsWriteEvidence(t, dir, "already-in-flight", "big-thinker", "feature-specialist", []string{"n/a"})

	out, code := runCent(t, bin, dir, "complete", "already-in-flight")
	if code == 0 {
		t.Fatalf("a partial legacy set must fail: %s", out)
	}
	if !strings.Contains(out, "big-thinker") || !strings.Contains(out, "feature-specialist") {
		t.Fatalf("message must name the legacy pair this workflow requires: %s", out)
	}
	if !strings.Contains(out, "planner-v1") {
		t.Fatalf("message must carry the contract annotation naming planner-v1: %s", out)
	}
}

// Scenario: evidence init for a legacy role succeeds on an unpinned workflow
func TestUPS_EvidenceInitSucceedsOnUnpinnedWorkflow(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)
	upsWriteWorkflow(t, dir, "already-in-flight", "")

	out, code := runCent(t, bin, dir, "evidence", "init", "already-in-flight", "big-thinker")
	if code != 0 {
		t.Fatalf("legacy workflow must still author big-thinker: %s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".workflow", "already-in-flight-big-thinker.json")); err != nil {
		t.Fatalf("stub not written: %v", err)
	}
}
