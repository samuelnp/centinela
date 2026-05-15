package worktree_test

import (
	"os/exec"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func TestRemove_MissingWorktreeIsNoOp(t *testing.T) {
	repo := initRepoForWorktrees(t)
	if err := worktree.Remove(repo, "ghost", false); err != nil {
		t.Fatalf("Remove on missing worktree must be a no-op: %v", err)
	}
}

func TestRemove_AfterCreate_RemovesDir(t *testing.T) {
	repo := initRepoForWorktrees(t)
	if _, err := worktree.Create(repo, "alpha"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := worktree.Remove(repo, "alpha", false); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if worktree.Exists(repo, "alpha") {
		t.Fatal("worktree should be gone after Remove")
	}
}

func TestDeleteBranch_AfterRemove_DeletesBranch(t *testing.T) {
	repo := initRepoForWorktrees(t)
	if _, err := worktree.Create(repo, "alpha"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := worktree.Remove(repo, "alpha", false); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if err := worktree.DeleteBranch(repo, "alpha"); err != nil {
		t.Fatalf("DeleteBranch: %v", err)
	}
	out, _ := exec.Command("git", "-C", repo, "rev-parse", "--verify", "refs/heads/alpha").CombinedOutput()
	if len(out) > 0 && string(out[:len(out)-1]) != "" {
		// rev-parse should fail for a missing branch; non-empty output here would
		// indicate the branch survived. Match by ensuring the next assertion fires.
	}
	if err := worktree.DeleteBranch(repo, "alpha"); err == nil {
		t.Fatal("second DeleteBranch should fail because branch is already gone")
	}
}
