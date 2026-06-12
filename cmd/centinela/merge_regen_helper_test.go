package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// seedCleanMergeRepo provisions a git repo with a committed feature worktree so
// runMerge takes the clean-merge path that fires docsPortalRegen. Returns the
// repo root (cwd is set to it).
func seedCleanMergeRepo(t *testing.T, feature string) string {
	t.Helper()
	d := t.TempDir()
	git := func(dir string, args ...string) {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git(d, "init", "-q", "-b", "main")
	git(d, "config", "user.email", "qa@centinela.dev")
	git(d, "config", "user.name", "QA")
	if err := os.WriteFile(filepath.Join(d, ".gitignore"), []byte(".worktrees/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "centinela.toml"),
		[]byte("[validate]\ncommands = []\n[gates]\nfile_size = false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(d, "add", ".")
	git(d, "commit", "-q", "-m", "seed")

	o, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(o) }) //nolint:errcheck
	os.Chdir(d)                       //nolint:errcheck

	wt := filepath.Join(d, ".worktrees", feature)
	git(d, "worktree", "add", ".worktrees/"+feature, "-b", feature)
	if err := os.WriteFile(filepath.Join(wt, "feature.txt"), []byte(feature+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(wt, "add", ".")
	git(wt, "commit", "-q", "-m", feature+" commit")
	return d
}
