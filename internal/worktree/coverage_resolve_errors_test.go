package worktree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func dirtyWorktree(t *testing.T, wt string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(wt, "untracked.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("dirty worktree: %v", err)
	}
}

// Merge surfaces the Remove failure when the clean-merged worktree still holds
// untracked files (`git worktree remove` refuses without --force).
func TestMerge_CleanButRemoveFails(t *testing.T) {
	repo, wt := setupMergeRepo(t, "omega")
	dirtyWorktree(t, wt)
	run := func(string) (bool, string) { return true, "" }
	if _, err := worktree.Merge(repo, "omega", run); err == nil {
		t.Fatal("Merge must surface the worktree removal failure")
	}
}

// ResolveMerge surfaces the Remove failure on an APPLY verdict when the
// worktree cannot be removed (untracked files block removal without --force).
func TestResolveMerge_RemoveFailureSurfaced(t *testing.T) {
	repo := resolveRepo(t, "rho2")
	writeMarker(t, repo, "rho2")
	dirtyWorktree(t, filepath.Join(repo, ".worktrees", "rho2"))
	apply := func(string) (string, error) { return "complete", nil }
	if _, err := worktree.ResolveMerge(repo, "rho2", apply); err == nil {
		t.Fatal("ResolveMerge must surface the Remove failure")
	}
}

// ResolveMerge surfaces the ClearPending failure when the worktree removes
// cleanly but the read-only .workflow dir blocks marker deletion.
func TestResolveMerge_ClearPendingFailureSurfaced(t *testing.T) {
	skipIfRoot(t)
	repo := resolveRepo(t, "sigma")
	writeMarker(t, repo, "sigma")
	wf := filepath.Join(repo, ".workflow")
	if err := os.Chmod(wf, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	defer os.Chmod(wf, 0o755)
	apply := func(string) (string, error) { return "complete", nil }
	if _, err := worktree.ResolveMerge(repo, "sigma", apply); err == nil {
		t.Fatal("ResolveMerge must surface the ClearPending failure")
	}
}

// WritePending surfaces the temp-write failure when .workflow is read-only.
func TestWritePending_TempWriteFails(t *testing.T) {
	skipIfRoot(t)
	repo := t.TempDir()
	wf := filepath.Join(repo, ".workflow")
	if err := os.Mkdir(wf, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Chmod(wf, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	defer os.Chmod(wf, 0o755)
	if err := worktree.WritePending(repo, worktree.MergeOutcome{Feature: "tau", TextConflict: true}); err == nil {
		t.Fatal("WritePending must error when the temp file cannot be written")
	}
}

// WritePending surfaces the rename failure when the final marker path is a
// directory (atomic rename cannot replace a directory with a file).
func TestWritePending_RenameFails(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".workflow"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Mkdir(worktree.PendingPath(repo, "upsilon"), 0o755); err != nil {
		t.Fatalf("mkdir marker dir: %v", err)
	}
	if err := worktree.WritePending(repo, worktree.MergeOutcome{Feature: "upsilon", TextConflict: true}); err == nil {
		t.Fatal("WritePending must error when the rename target is a directory")
	}
}
