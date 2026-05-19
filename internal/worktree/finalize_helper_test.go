package worktree_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// resolveRepo builds a hermetic repo with one committed worktree for the
// feature so ResolveMerge's clean-tree re-check and Remove path are exercised
// against real git. Returns the repo root.
func resolveRepo(t *testing.T, feature string) string {
	t.Helper()
	d := t.TempDir()
	git := func(dir string, args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	for _, a := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.email", "qa@centinela.dev"},
		{"config", "user.name", "QA"},
	} {
		git(d, a...)
	}
	_ = os.WriteFile(filepath.Join(d, ".gitignore"),
		[]byte(".worktrees/\n.workflow/\n"), 0o644)
	git(d, "add", ".")
	git(d, "commit", "-q", "-m", "seed")
	wt := filepath.Join(d, ".worktrees", feature)
	git(d, "worktree", "add", filepath.Join(".worktrees", feature), "-b", feature)
	_ = os.WriteFile(filepath.Join(wt, "f.txt"), []byte(feature+"\n"), 0o644)
	git(wt, "add", ".")
	git(wt, "commit", "-q", "-m", "feature commit")
	return d
}

func writeMarker(t *testing.T, repo, feature string) {
	t.Helper()
	o := worktree.MergeOutcome{Feature: feature, TextConflict: true}
	if err := worktree.WritePending(repo, o); err != nil {
		t.Fatalf("WritePending: %v", err)
	}
}

func okValidator(verdict string) worktree.StewardEvidenceValidator {
	return func(string) (string, error) { return verdict, nil }
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
