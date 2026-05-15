package worktree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func TestMerge_TextConflict_KeepsWorktreeAndReportsPaths(t *testing.T) {
	repo := initRepoForWorktrees(t)
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
