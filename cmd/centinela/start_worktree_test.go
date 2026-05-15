package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunStart_WithUseWorktrees_ProvisionsWorktreeAndStoresPath(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck

	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.email", "qa@centinela.dev"},
		{"config", "user.name", "QA"},
	} {
		c := exec.Command("git", args...)
		c.Dir = d
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	_ = os.WriteFile(filepath.Join(d, "PROJECT.md"), []byte("Project Stage: existing\n"), 0644)
	_ = os.WriteFile(filepath.Join(d, ".gitignore"), []byte(".worktrees/\n"), 0644)
	_ = os.WriteFile(filepath.Join(d, "centinela.toml"),
		[]byte("[workflow]\nuse_worktrees = true\n"), 0644)
	for _, args := range [][]string{{"add", "."}, {"commit", "-q", "-m", "seed"}} {
		c := exec.Command("git", args...)
		c.Dir = d
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	os.Chdir(d) //nolint:errcheck
	if err := runStart(nil, []string{"alpha"}); err != nil {
		t.Fatalf("runStart: %v", err)
	}
	if _, err := os.Stat(filepath.Join(d, ".worktrees", "alpha")); err != nil {
		t.Fatalf("worktree not created: %v", err)
	}
}
