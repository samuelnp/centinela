package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// setupMergeRepo builds: main with one commit, a feature worktree with a
// non-conflicting commit on a new branch. Caller picks the feature name.
func setupMergeRepo(t *testing.T, feature string) (repo, wt string) {
	t.Helper()
	repo = initRepoForWorktrees(t)
	var err error
	wt, err = worktree.Create(repo, feature)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// Add a unique file inside the worktree so the branch has new content.
	path := filepath.Join(wt, "feat-"+feature+".txt")
	if err := os.WriteFile(path, []byte("from "+feature), 0644); err != nil {
		t.Fatalf("write feature file: %v", err)
	}
	gitRun(t, wt, "add", ".")
	gitRun(t, wt, "commit", "-q", "-m", "feature commit "+feature)
	return repo, wt
}

func TestMerge_CleanRemovesWorktree(t *testing.T) {
	repo, wt := setupMergeRepo(t, "gamma")
	passing := func(_ string) (bool, string) { return true, "ok" }
	out, err := worktree.Merge(repo, "gamma", passing)
	if err != nil {
		t.Fatalf("Merge clean: %v", err)
	}
	if out.TextConflict || out.ValidateFail {
		t.Fatalf("unexpected flags on clean merge: %+v", out)
	}
	if out.WorktreeKept {
		t.Fatal("WorktreeKept should be false on clean success")
	}
	if _, err := os.Stat(wt); !os.IsNotExist(err) {
		t.Fatalf("worktree dir should be removed; err=%v", err)
	}
}

func TestMerge_TextConflict_KeepsWorktreeAndReportsPaths(t *testing.T) {
	repo := initRepoForWorktrees(t)
	// main edits shared.txt; feature edits the same file → conflict on merge.
	shared := filepath.Join(repo, "shared.txt")
	if err := os.WriteFile(shared, []byte("base\n"), 0644); err != nil {
		t.Fatalf("write base: %v", err)
	}
	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-q", "-m", "base shared")
	wt, err := worktree.Create(repo, "delta")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wt, "shared.txt"), []byte("feature\n"), 0644); err != nil {
		t.Fatalf("write feature shared: %v", err)
	}
	gitRun(t, wt, "add", ".")
	gitRun(t, wt, "commit", "-q", "-m", "feature shared")
	// Re-edit shared on main so the histories diverge.
	if err := os.WriteFile(shared, []byte("main edit\n"), 0644); err != nil {
		t.Fatalf("write main edit: %v", err)
	}
	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-q", "-m", "main edit shared")

	passing := func(_ string) (bool, string) { return true, "" }
	out, err := worktree.Merge(repo, "delta", passing)
	if err != nil {
		t.Fatalf("Merge with text conflict should not error: %v", err)
	}
	if !out.TextConflict {
		t.Fatalf("expected TextConflict=true, got %+v", out)
	}
	if !out.WorktreeKept {
		t.Fatal("worktree must be kept on text conflict")
	}
	if len(out.ConflictedPaths) == 0 {
		t.Fatal("ConflictedPaths must enumerate conflicted files")
	}
	if reason := out.StewardReason(); reason != "git-text-conflict" {
		t.Fatalf("StewardReason = %q, want git-text-conflict", reason)
	}
}

func TestMerge_ValidateFail_KeepsWorktree(t *testing.T) {
	repo, wt := setupMergeRepo(t, "epsilon")
	failing := func(_ string) (bool, string) { return false, "test suite failed" }
	out, err := worktree.Merge(repo, "epsilon", failing)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if out.TextConflict {
		t.Fatalf("git merge applied cleanly; TextConflict should be false")
	}
	if !out.ValidateFail {
		t.Fatal("ValidateFail must be true when validator returns false")
	}
	if !out.WorktreeKept {
		t.Fatal("worktree must be kept when validate fails")
	}
	if _, err := os.Stat(wt); err != nil {
		t.Fatalf("worktree should remain on disk: %v", err)
	}
	if out.StewardReason() != "post-merge-validate-failed" {
		t.Fatalf("StewardReason = %q", out.StewardReason())
	}
}

func TestMerge_DirtyMain_FailsFast(t *testing.T) {
	repo, _ := setupMergeRepo(t, "kappa")
	// Dirty main with an unstaged change.
	if err := os.WriteFile(filepath.Join(repo, "dirty.txt"), []byte("oops"), 0644); err != nil {
		t.Fatalf("write dirty: %v", err)
	}
	gitRun(t, repo, "add", "dirty.txt")
	called := false
	runner := func(_ string) (bool, string) { called = true; return true, "" }
	_, err := worktree.Merge(repo, "kappa", runner)
	if err == nil {
		t.Fatal("Merge must fail fast on dirty main")
	}
	if called {
		t.Fatal("validator must not run when pre-check fails")
	}
	// Worktree was left in place for the user.
	if _, err := os.Stat(filepath.Join(repo, ".worktrees", "kappa")); err != nil {
		t.Fatalf("worktree should be untouched after dirty pre-check: %v", err)
	}
}

// Ensure the unused exec import stays in case of platform-specific assertions.
var _ = exec.Command
