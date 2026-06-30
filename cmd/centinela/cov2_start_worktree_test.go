package main

import (
	"os"
	"os/exec"
	"testing"
)

// TestCov2RunStartWorktreeProvisionError drives runStart's MaybeProvision error
// branch: worktrees are enabled and the directory is a git repo, but worktree
// creation fails because the repository has no commits (unborn HEAD).
func TestCov2RunStartWorktreeProvisionError(t *testing.T) {
	d := t.TempDir()
	if out, err := exec.Command("git", "-C", d, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s", out)
	}
	if err := os.WriteFile(d+"/PROJECT.md", []byte("Project Stage: existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(d+"/centinela.toml", []byte("[workflow]\nuse_worktrees = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d)
	if err := runStart(nil, []string{"okfeat"}); err == nil {
		t.Fatal("expected a worktree provisioning error on an unborn repo")
	}
}
