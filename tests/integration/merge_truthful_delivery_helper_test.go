package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// mtdiGit runs git in dir, failing the test on error, and returns stdout.
func mtdiGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// mtdiRepo builds a committed repo plus a feature worktree with one commit
// on the feature branch that main does not have.
func mtdiRepo(t *testing.T, feature string) (repo, wt string) {
	t.Helper()
	repo = t.TempDir()
	mtdiGit(t, repo, "init", "-q", "-b", "main")
	mtdiGit(t, repo, "config", "user.email", "qa@centinela.dev")
	mtdiGit(t, repo, "config", "user.name", "QA")
	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte(".worktrees/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mtdiGit(t, repo, "add", ".")
	mtdiGit(t, repo, "commit", "-q", "-m", "seed")
	var err error
	wt, err = worktree.Create(repo, feature)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wt, "feat.txt"), []byte(feature+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mtdiGit(t, wt, "add", ".")
	mtdiGit(t, wt, "commit", "-q", "-m", "feature commit")
	return repo, wt
}
