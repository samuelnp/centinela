package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: merge from inside the feature worktree advances main and removes the worktree
func TestAccDeliverMergeFromWorktreeCwdAdvancesMain(t *testing.T) {
	bin := buildCent(t)
	repo := mergeRepo(t, "high-score", false)
	wt := filepath.Join(repo, ".worktrees", "high-score")
	cdpWorkflow(t, wt, "high-score", false, true) // deliver reads state from its CWD
	before := mtdaGit(t, repo, "rev-parse", "main")
	out, code := runCent(t, bin, wt, "deliver", "high-score", "--via", "merge")
	if code != 0 {
		t.Fatalf("deliver --via merge from worktree CWD must exit 0 (code=%d):\n%s", code, out)
	}
	if mtdaGit(t, repo, "rev-parse", "main") == before {
		t.Fatalf("main did NOT advance in the primary tree — false success:\n%s", out)
	}
	mtdaGit(t, repo, "merge-base", "--is-ancestor", "high-score", "main")
	if _, e := os.Stat(wt); !os.IsNotExist(e) {
		t.Fatalf("worktree must be removed; err=%v\n%s", e, out)
	}
	if !strings.Contains(out, "high-score") {
		t.Fatalf("success message must name the feature:\n%s", out)
	}
}

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: merge run from inside a worktree prints a one-line primary-tree notice
func TestAccMergeInsideWorktreePrintsPrimaryNotice(t *testing.T) {
	bin := buildCent(t)
	repo := mergeRepo(t, "high-score", false)
	wt := filepath.Join(repo, ".worktrees", "high-score")
	out, code := runCent(t, bin, wt, "merge", "high-score")
	if code != 0 {
		t.Fatalf("merge from worktree CWD must exit 0 (code=%d):\n%s", code, out)
	}
	if !strings.Contains(out, "Merging in primary working tree") {
		t.Fatalf("want the one-line primary-tree notice:\n%s", out)
	}
}

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: merge run from the primary tree behaves as before, with verification applied
func TestAccMergeFromPrimaryTreeNoNotice(t *testing.T) {
	bin := buildCent(t)
	repo := mergeRepo(t, "high-score", false)
	before := mtdaGit(t, repo, "rev-parse", "main")
	out, code := runCent(t, bin, repo, "merge", "high-score")
	if code != 0 {
		t.Fatalf("merge from primary tree must exit 0 (code=%d):\n%s", code, out)
	}
	if mtdaGit(t, repo, "rev-parse", "main") == before {
		t.Fatalf("main must advance:\n%s", out)
	}
	if _, e := os.Stat(filepath.Join(repo, ".worktrees", "high-score")); !os.IsNotExist(e) {
		t.Fatalf("worktree must be removed; err=%v", e)
	}
	if strings.Contains(out, "Merging in primary working tree") {
		t.Fatalf("no worktree notice when already in the primary tree:\n%s", out)
	}
}
