package unit_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// mtdRepo builds a minimal committed git repo for PrimaryTree resolution.
func mtdRepo(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	git := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = d
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q", "-b", "main")
	git("config", "user.email", "qa@centinela.dev")
	git("config", "user.name", "QA")
	if err := os.WriteFile(filepath.Join(d, ".gitignore"), []byte(".worktrees/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-q", "-m", "seed")
	return d
}

// PrimaryTree resolves the SAME primary checkout whether asked from the repo
// root or from inside a feature worktree — the property the merge relies on.
func TestPrimaryTreeResolvesIdenticallyFromRootAndWorktree(t *testing.T) {
	repo := mtdRepo(t)
	wt, err := worktree.Create(repo, "alpha")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	fromRoot, err := worktree.PrimaryTree(repo)
	if err != nil {
		t.Fatalf("PrimaryTree from root: %v", err)
	}
	fromWt, err := worktree.PrimaryTree(wt)
	if err != nil {
		t.Fatalf("PrimaryTree from worktree: %v", err)
	}
	want, _ := filepath.EvalSymlinks(repo)
	gotRoot, _ := filepath.EvalSymlinks(fromRoot)
	gotWt, _ := filepath.EvalSymlinks(fromWt)
	if gotRoot != want || gotWt != want {
		t.Fatalf("PrimaryTree mismatch: root=%q worktree=%q want=%q", gotRoot, gotWt, want)
	}
}

// Outside any git repository PrimaryTree refuses instead of guessing.
func TestPrimaryTreeRefusesOutsideAnyRepo(t *testing.T) {
	_, err := worktree.PrimaryTree(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "cannot resolve primary working tree") {
		t.Fatalf("want the never-guess refusal, got: %v", err)
	}
}
