package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: removal is only claimed when the worktree directory is actually gone
//
// A locked worktree survives `git worktree remove` (even with --force), so the
// merge lands but removal does not. Nothing may claim it was removed.
func TestAccMergeSurvivingWorktreeIsNeverClaimedRemoved(t *testing.T) {
	bin := buildCent(t)
	repo := mergeRepo(t, "high-score", false)
	wt := filepath.Join(repo, ".worktrees", "high-score")
	mtdaGit(t, repo, "worktree", "lock", wt)
	before := mtdaGit(t, repo, "rev-parse", "main")

	out, code := runCent(t, bin, repo, "merge", "high-score")
	if code == 0 {
		t.Fatalf("a surviving worktree must exit non-zero (code=%d):\n%s", code, out)
	}
	if strings.Contains(out, mtdcSuccess) {
		t.Fatalf("FALSE CLAIM — the worktree is still on disk:\n%s", out)
	}
	if _, e := os.Stat(wt); e != nil {
		t.Fatalf("precondition: the locked worktree must have survived: %v", e)
	}
	if reg := mtdaGit(t, repo, "worktree", "list", "--porcelain"); !strings.Contains(reg, wt) {
		t.Fatalf("precondition: the worktree must still be registered:\n%s", reg)
	}
	if mtdaGit(t, repo, "rev-parse", "main") == before {
		t.Fatalf("precondition: the merge itself should have landed:\n%s", out)
	}
	if !strings.Contains(out, "worktree removal failed") {
		t.Fatalf("the failure must name the outstanding step:\n%s", out)
	}
}

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: the success message is never printed when the ref did not advance
//
// Real git cannot produce "merge exits 0, HEAD unmoved, branch NOT an
// ancestor" — the conditions are coupled. A `git` shim that no-ops only
// `merge --no-ff` manufactures it, and the shipped guard is exercised
// end-to-end through the real binary.
func TestAccMergeNoRefAdvanceNeverPrintsSuccess(t *testing.T) {
	bin := buildCent(t)
	repo := mergeRepo(t, "high-score", false)
	env := mtdfNoOpMergeGit(t)
	before := mtdaGit(t, repo, "rev-parse", "main")

	out, code := runCentEnv(t, bin, repo, env, "merge", "high-score")
	if code == 0 {
		t.Fatalf("a merge that did not deliver must exit non-zero (code=%d):\n%s", code, out)
	}
	if !strings.Contains(out, "did not advance") {
		t.Fatalf("the refusal must say the ref did not advance:\n%s", out)
	}
	if strings.Contains(out, `Merged "high-score" into main`) ||
		strings.Contains(out, "already merged") {
		t.Fatalf("FABRICATED SUCCESS on an unmoved ref:\n%s", out)
	}
	if mtdaGit(t, repo, "rev-parse", "main") != before {
		t.Fatal("precondition: the shim must leave main unmoved")
	}
	if _, e := os.Stat(filepath.Join(repo, ".worktrees", "high-score")); e != nil {
		t.Fatalf("the worktree must survive an unverified merge: %v", e)
	}
}
