// Acceptance: specs/adversarial-validate-verifier.feature
package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: Revision skew after fixes landed on top of the verified commit
func TestAVV_RevisionSkewDemandsFreshVerification(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "post-fix-drift", "adversarial-v1")
	avvWriteReport(t, dir, "post-fix-drift", avvReport("SAFE", "", avvGroundedCommands))
	avvStamp(t, bin, dir, "post-fix-drift")

	mustWrite(t, filepath.Join(dir, "src.go"), "package x\n\nfunc fix() {}\n")
	commit(t, dir, "fix landed on top of the verified commit")

	out, code := avvComplete(t, bin, dir, "post-fix-drift")
	if code == 0 {
		t.Fatalf("revision skew must block, got exit 0: %s", out)
	}
	if !strings.Contains(out, "stale") || !strings.Contains(out, "FRESH verifier") {
		t.Fatalf("message must say stale and demand a fresh verifier: %s", out)
	}
}

// Scenario: Uncommitted in-place fixes leave HEAD unchanged but stale the digest
func TestAVV_UncommittedPatchStalesTheDigest(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "uncommitted-patch", "adversarial-v1")
	avvWriteReport(t, dir, "uncommitted-patch", avvReport("SAFE", "", avvGroundedCommands))
	avvStamp(t, bin, dir, "uncommitted-patch")

	// Same HEAD, no commit — a HEAD-only comparison would wrongly admit this.
	mustWrite(t, filepath.Join(dir, "src.go"), "package x\n\nfunc uncommitted() {}\n")

	out, code := avvComplete(t, bin, dir, "uncommitted-patch")
	if code == 0 {
		t.Fatalf("uncommitted tracked-file edit must block, got exit 0: %s", out)
	}
	if !strings.Contains(out, "stale") || !strings.Contains(out, "uncommitted") {
		t.Fatalf("message must say stale due to uncommitted changes: %s", out)
	}
}

// Scenario: Mutating only files under .workflow/ does not stale a fresh verification (D3a)
func TestAVV_WorkflowOnlyChurnSurvivesFreshness(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "workflow-only-churn", "adversarial-v1")
	avvWriteReport(t, dir, "workflow-only-churn", avvReport("SAFE", "", avvGroundedCommands))
	avvStamp(t, bin, dir, "workflow-only-churn")

	// The verifier's own report write must never self-invalidate its stamp.
	// (Deliberately NOT named "*-gatekeeper.json" — that path IS the role's
	// evidence file and would make verify.Verify run real claim checks.)
	mustWrite(t, filepath.Join(dir, ".workflow", "workflow-only-churn-scratch.txt"), "noise")

	out, code := avvComplete(t, bin, dir, "workflow-only-churn")
	if code != 0 {
		t.Fatalf(".workflow/-only churn must not stale the verification: %d: %s", code, out)
	}
}

// Scenario: centinela artifact stamp records the verified revision and tree digest
func TestAVV_ArtifactStampRecordsRevisionAndPreservesCommands(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "stamp-me", "adversarial-v1")
	avvWriteReport(t, dir, "stamp-me", avvReport("SAFE", "", avvGroundedCommands))

	out, code := runCent(t, bin, dir, "artifact", "stamp", "stamp-me")
	if code != 0 {
		t.Fatalf("artifact stamp must succeed: %s", out)
	}
	if !strings.Contains(out, "revision:") || !strings.Contains(out, "treeDigest:") {
		t.Fatalf("stamp output must report the recorded revision and treeDigest: %s", out)
	}

	body := readFile(t, filepath.Join(dir, ".workflow", "stamp-me-gatekeeper.md"))
	if strings.Contains(body, `"revision":""`) || strings.Contains(body, `"treeDigest":""`) {
		t.Fatalf("stamp must have filled revision/treeDigest in the report: %s", body)
	}
	if !strings.Contains(body, `"argv":["centinela","validate"]`) {
		t.Fatalf("stamp must leave the commands array byte-identical: %s", body)
	}
}
