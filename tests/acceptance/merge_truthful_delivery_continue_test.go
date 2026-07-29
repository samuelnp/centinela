package acceptance_test

import (
	"os"
	"strings"
	"testing"
)

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: merge --continue never claims success while the target ref is unmoved
//
// The regression this feature exists to kill, in its last hiding place: a
// stall started from a worktree CWD, resumed from a worktree CWD, with the
// merge NOT actually resolved. It used to exit 0 printing
// `Merged "high-score" into main and removed its worktree.`
func TestAccMergeContinueFromWorktreeCwdNeverFakesSuccess(t *testing.T) {
	bin := buildCent(t)
	repo, wt, before := mtdcStall(t, bin, "high-score", true)
	writeMergeEvidence(t, repo, "high-score", "complete")
	out, code := runCent(t, bin, wt, "merge", "--continue", "high-score")
	if code == 0 {
		t.Fatalf("an unresolved merge must not resume successfully:\n%s", out)
	}
	if strings.Contains(out, mtdcSuccess) {
		t.Fatalf("FABRICATED SUCCESS — main never moved:\n%s", out)
	}
	if mtdaGit(t, repo, "rev-parse", "main") != before {
		t.Fatalf("main must be untouched by the refusal:\n%s", out)
	}
	if _, e := os.Stat(wt); e != nil {
		t.Fatalf("worktree must survive the refusal: %v", e)
	}
}

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: a merge stalled from a worktree CWD is resumable from that same worktree CWD
func TestAccMergeContinueResumesFromWorktreeCwd(t *testing.T) {
	bin := buildCent(t)
	repo, wt, before := mtdcStall(t, bin, "high-score", true)
	applyRepoMergeResolution(t, repo)
	writeMergeEvidence(t, repo, "high-score", "complete")
	out, code := runCent(t, bin, wt, "merge", "--continue", "high-score")
	if code != 0 {
		t.Fatalf("a resolved merge must resume from the worktree CWD (code=%d):\n%s", code, out)
	}
	if mtdaGit(t, repo, "rev-parse", "main") == before {
		t.Fatalf("main must have advanced:\n%s", out)
	}
	mtdcAssertDelivered(t, repo, wt, "high-score", out)
}

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: a merge stalled from a worktree CWD is resumable from the primary CWD
func TestAccMergeContinueResumesFromPrimaryCwd(t *testing.T) {
	bin := buildCent(t)
	repo, wt, _ := mtdcStall(t, bin, "high-score", true)
	applyRepoMergeResolution(t, repo)
	writeMergeEvidence(t, repo, "high-score", "complete")
	out, code := runCent(t, bin, repo, "merge", "--continue", "high-score")
	if code != 0 {
		t.Fatalf("the same stall must resume from the primary CWD (code=%d):\n%s", code, out)
	}
	mtdcAssertDelivered(t, repo, wt, "high-score", out)
}
