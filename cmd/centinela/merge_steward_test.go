package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunMerge_TextConflict_ReturnsStewardRequired(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck

	gitInit := func(dir string, args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.email", "qa@centinela.dev"},
		{"config", "user.name", "QA"},
	} {
		gitInit(d, args...)
	}
	_ = os.WriteFile(filepath.Join(d, "shared.txt"), []byte("base\n"), 0644)
	_ = os.WriteFile(filepath.Join(d, ".gitignore"), []byte(".worktrees/\n"), 0644)
	_ = os.WriteFile(filepath.Join(d, "centinela.toml"),
		[]byte("[validate]\ncommands = []\n[gates]\nfile_size = false\n"), 0644)
	gitInit(d, "add", ".")
	gitInit(d, "commit", "-q", "-m", "base")

	// Provision the worktree, edit shared.txt there, commit.
	wt := filepath.Join(d, ".worktrees", "delta")
	gitInit(d, "worktree", "add", ".worktrees/delta", "-b", "delta")
	_ = os.WriteFile(filepath.Join(wt, "shared.txt"), []byte("feature\n"), 0644)
	gitInit(wt, "add", ".")
	gitInit(wt, "commit", "-q", "-m", "feature edit")

	// Diverge main so the merge is a real text conflict.
	_ = os.WriteFile(filepath.Join(d, "shared.txt"), []byte("main\n"), 0644)
	gitInit(d, "add", ".")
	gitInit(d, "commit", "-q", "-m", "main edit")

	os.Chdir(d) //nolint:errcheck
	err := runMerge(nil, []string{"delta"})
	if err == nil {
		t.Fatal("runMerge should return an error when text-conflict invokes Steward")
	}
	if !strings.Contains(err.Error(), "Merge Steward review") {
		t.Fatalf("expected steward review hint, got: %v", err)
	}
	if _, err := os.Stat(filepath.Join(d, ".workflow", "delta-merge-pending.json")); err != nil {
		t.Fatalf("pending marker must be written on text conflict: %v", err)
	}
}
