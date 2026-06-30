package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCov2CommitStepReportsRealCommitFailure drives commitStep's stderr branch:
// `git add -A` succeeds but `git commit` fails for a reason other than
// "nothing to commit" (an unresolvable author identity, forced via a
// useConfigOnly global config with no name). commitStep must not panic.
func TestCov2CommitStepReportsRealCommitFailure(t *testing.T) {
	d := t.TempDir()
	gcfg := filepath.Join(d, "gitconfig")
	if err := os.WriteFile(gcfg, []byte("[user]\n\tuseConfigOnly = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", gcfg)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	t.Setenv("GIT_AUTHOR_NAME", "")
	t.Setenv("GIT_AUTHOR_EMAIL", "")
	t.Setenv("GIT_COMMITTER_NAME", "")
	t.Setenv("GIT_COMMITTER_EMAIL", "")

	if out, err := exec.Command("git", "-C", d, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s", out)
	}
	if err := os.WriteFile(filepath.Join(d, "f.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d)
	// add -A stages f.txt; commit then fails on the missing identity, taking
	// the stderr-reporting branch (message is not "nothing to commit").
	commitStep("feat", "plan", 1, 5)
}
