package worktree_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// Finding 3: an untracked file in the worktree (the normal state at delivery
// time) makes `git worktree remove` refuse AFTER main has already advanced.
// The outcome must carry both halves so the caller can say so, and the
// operator must have a way out.
func TestMerge_BusyWorktree_HalfSuccessThenForceRemoveRecovers(t *testing.T) {
	repo, wt := setupMergeRepo(t, "high-score")
	if err := os.WriteFile(filepath.Join(wt, "scratch.txt"), []byte("busy\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := gitOut(t, repo, "rev-parse", "main")
	out, err := worktree.Merge(repo, "high-score", passingValidate)
	if err == nil {
		t.Fatal("a worktree git refuses to remove must not be reported as removed")
	}
	if !out.RefAdvanced || !out.RemoveFailed {
		t.Fatalf("half-success must record BOTH the advance and the removal failure: %+v", out)
	}
	if gitOut(t, repo, "rev-parse", "main") == before {
		t.Fatal("precondition: the merge itself should have landed")
	}
	// Recovery: the re-run finds the merge already landed and, with force,
	// completes the only outstanding step.
	rec, err := worktree.Merge(repo, "high-score", passingValidate, worktree.WithForceRemove())
	if err != nil {
		t.Fatalf("--force-remove re-run must recover: %v", err)
	}
	if !rec.AlreadyMerged || rec.RemoveFailed {
		t.Fatalf("recovery run must report already-merged and a clean removal: %+v", rec)
	}
	if _, e := os.Stat(wt); !os.IsNotExist(e) {
		t.Fatalf("worktree must be gone after the forced retry; err=%v", e)
	}
}

// Finding 2 end-to-end: a worktree moved outside `.worktrees/<feature>` is
// removed for real (registry path), never merely claimed.
func TestMerge_WorktreeOutsideConvention_ReallyRemoved(t *testing.T) {
	repo, wt := setupMergeRepo(t, "high-score")
	elsewhere := filepath.Join(t.TempDir(), "relocated")
	gitRun(t, repo, "worktree", "move", wt, elsewhere)
	out, err := worktree.Merge(repo, "high-score", passingValidate)
	if err != nil {
		t.Fatalf("Merge with a relocated worktree: %v", err)
	}
	if !out.RefAdvanced {
		t.Fatalf("want a verified advance, got %+v", out)
	}
	if _, e := os.Stat(elsewhere); !os.IsNotExist(e) {
		t.Fatalf("the relocated worktree must actually be removed; err=%v", e)
	}
	if reg := gitOut(t, repo, "worktree", "list", "--porcelain"); strings.Contains(reg, elsewhere) {
		t.Fatalf("worktree must be gone from the registry:\n%s", reg)
	}
}
