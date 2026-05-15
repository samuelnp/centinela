package worktree_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func initRepoForWorktrees(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q", "-b", "main")
	gitRun(t, dir, "config", "user.email", "qa@centinela.dev")
	gitRun(t, dir, "config", "user.name", "QA")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("seed\n"), 0644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(".worktrees/\n"), 0644); err != nil {
		t.Fatalf("write gitignore: %v", err)
	}
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-q", "-m", "seed")
	return dir
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestCreate_FreshWorktreeAndBranch(t *testing.T) {
	repo := initRepoForWorktrees(t)
	path, err := worktree.Create(repo, "alpha")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("worktree dir not on disk: %v", err)
	}
	if out, err := exec.Command("git", "-C", repo, "rev-parse", "--verify", "refs/heads/alpha").CombinedOutput(); err != nil {
		t.Fatalf("branch alpha missing: %v\n%s", err, out)
	}
}

func TestCreate_Idempotent_SecondCallNoOp(t *testing.T) {
	repo := initRepoForWorktrees(t)
	first, err := worktree.Create(repo, "alpha")
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}
	second, err := worktree.Create(repo, "alpha")
	if err != nil {
		t.Fatalf("second Create must succeed: %v", err)
	}
	if first != second {
		t.Fatalf("Create not idempotent: %q vs %q", first, second)
	}
}

func TestCreate_ReusesExistingBranch(t *testing.T) {
	repo := initRepoForWorktrees(t)
	gitRun(t, repo, "branch", "alpha")
	path, err := worktree.Create(repo, "alpha")
	if err != nil {
		t.Fatalf("Create with existing branch failed: %v", err)
	}
	out, err := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD in worktree: %v", err)
	}
	if got := string(out); got == "" || got[:5] != "alpha" {
		t.Fatalf("worktree not on alpha branch: %q", got)
	}
}

func TestCreate_InvalidSlug_RefusesBeforeDiskTouch(t *testing.T) {
	repo := initRepoForWorktrees(t)
	if _, err := worktree.Create(repo, "alpha/../beta"); err == nil {
		t.Fatal("Create accepted path-escape slug")
	}
	if _, err := os.Stat(filepath.Join(repo, ".worktrees")); !os.IsNotExist(err) {
		t.Fatalf(".worktrees should not exist after rejected slug: err=%v", err)
	}
}
