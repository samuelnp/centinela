package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

// Create surfaces a git error when the target branch is already checked out
// in the main working tree (git refuses a second checkout of the same branch).
func TestCreate_BranchCheckedOutElsewhere_Errors(t *testing.T) {
	repo := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		if out, err := gitRunner(repo, args...); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	run("config", "user.email", "qa@centinela.dev")
	run("config", "user.name", "QA")
	if err := os.WriteFile(filepath.Join(repo, "f.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	run("add", ".")
	run("commit", "-q", "-m", "seed")
	// "main" is the currently checked-out branch in the primary worktree.
	// Creating a worktree on a branch named "main" must fail.
	if _, err := Create(repo, "main"); err == nil {
		t.Fatal("Create must error when the branch is already checked out in the primary tree")
	}
}
