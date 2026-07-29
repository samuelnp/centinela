package worktree_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// Re-running Merge for a branch that already landed reports AlreadyMerged —
// never RefAdvanced, never an error — and leaves main exactly where it was.
func TestMerge_AlreadyMergedRerun_HonestAndIdempotent(t *testing.T) {
	repo, _ := setupMergeRepo(t, "high-score")
	if _, err := worktree.Merge(repo, "high-score", passingValidate); err != nil {
		t.Fatalf("first Merge: %v", err)
	}
	before := gitOut(t, repo, "rev-parse", "main")
	out, err := worktree.Merge(repo, "high-score", passingValidate)
	if err != nil {
		t.Fatalf("already-merged rerun must succeed honestly: %v", err)
	}
	if out.RefAdvanced || !out.AlreadyMerged {
		t.Fatalf("want AlreadyMerged only on rerun, got %+v", out)
	}
	if gitOut(t, repo, "rev-parse", "main") != before {
		t.Fatal("rerun must not move main")
	}
}

// A real detached HEAD in the primary tree is refused before any merge.
func TestMerge_DetachedPrimaryHead_RealGitRefused(t *testing.T) {
	repo, wt := setupMergeRepo(t, "high-score")
	gitRun(t, repo, "checkout", "-q", "--detach")
	before := gitOut(t, repo, "rev-parse", "HEAD")
	_, err := worktree.Merge(repo, "high-score", passingValidate)
	if err == nil || !strings.Contains(err.Error(), "detached HEAD") {
		t.Fatalf("want detached-HEAD refusal, got: %v", err)
	}
	if gitOut(t, repo, "rev-parse", "HEAD") != before {
		t.Fatal("refusal must not move HEAD")
	}
	if _, err := os.Stat(wt); err != nil {
		t.Fatalf("worktree must survive the refusal: %v", err)
	}
}
