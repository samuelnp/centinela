package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: re-running deliver for an already-merged branch reports honestly
func TestAccDeliverAlreadyMergedReportsHonestly(t *testing.T) {
	bin := buildCent(t)
	repo := mergeRepo(t, "high-score", false)
	if out, code := runCent(t, bin, repo, "merge", "high-score"); code != 0 {
		t.Fatalf("first merge must land (code=%d):\n%s", code, out)
	}
	// Recreate the worktree for the already-landed branch: cleanup must
	// still run on the already-merged path.
	mtdaGit(t, repo, "worktree", "add", filepath.Join(".worktrees", "high-score"), "high-score")
	wt := filepath.Join(repo, ".worktrees", "high-score")
	cdpWorkflow(t, repo, "high-score", false, true)
	before := mtdaGit(t, repo, "rev-parse", "main")
	out, code := runCent(t, bin, repo, "deliver", "high-score", "--via", "merge")
	if code != 0 {
		t.Fatalf("already-merged deliver must exit 0 (code=%d):\n%s", code, out)
	}
	if !strings.Contains(out, "already merged") {
		t.Fatalf("want the honest already-merged wording:\n%s", out)
	}
	if strings.Contains(out, `Merged "high-score" into main`) {
		t.Fatalf("must not fabricate a fresh delivery:\n%s", out)
	}
	if mtdaGit(t, repo, "rev-parse", "main") != before {
		t.Fatal("already-merged deliver must not move main")
	}
	if _, e := os.Stat(wt); !os.IsNotExist(e) {
		t.Fatalf("cleanup must still run on the already-merged path; err=%v", e)
	}
}

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: already-merged re-run with the worktree already gone is an idempotent honest success
func TestAccMergeRerunWorktreeGoneIdempotent(t *testing.T) {
	bin := buildCent(t)
	repo := mergeRepo(t, "high-score", false)
	if out, code := runCent(t, bin, repo, "merge", "high-score"); code != 0 {
		t.Fatalf("first merge must land (code=%d):\n%s", code, out)
	}
	out, code := runCent(t, bin, repo, "merge", "high-score") // worktree already gone
	if code != 0 {
		t.Fatalf("idempotent rerun must exit 0 (code=%d):\n%s", code, out)
	}
	if !strings.Contains(out, "already merged") {
		t.Fatalf("want the honest already-merged wording:\n%s", out)
	}
}
