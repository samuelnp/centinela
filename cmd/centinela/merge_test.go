package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunMerge_InvalidSlug_Errors(t *testing.T) {
	err := runMerge(nil, []string{"Alpha/../beta"})
	if err == nil {
		t.Fatal("expected error for invalid slug")
	}
}

func TestRunMerge_SpecConflict_Blocks(t *testing.T) {
	// A real git repo is required: runMerge now resolves the primary working
	// tree BEFORE spec-conflict detection and refuses outside any repository.
	d := seedCleanMergeRepo(t, "zeta") // cwd = d

	// Parallel divergence: two in-flight worktrees carry the SAME spec file
	// with the SAME scenario resolved differently. Narrowed by the
	// spec-conflict-false-positives hotfix — main-vs-branch differences and
	// differently-named files are supersession, not conflicts.
	spec := func(feat, then string) {
		dir := filepath.Join(d, ".worktrees", feat, "specs")
		_ = os.MkdirAll(dir, 0755)
		_ = os.WriteFile(filepath.Join(dir, "login.feature"),
			[]byte("Feature: L\n  Scenario: clash\n    Given ctx\n    Then "+then+"\n"), 0644)
	}
	spec("zeta", "A")
	spec("eta", "B")

	err := runMerge(nil, []string{"zeta"})
	if err == nil {
		t.Fatal("expected spec-conflict error")
	}
	if !strings.Contains(err.Error(), "spec conflicts") {
		t.Fatalf("error should mention spec conflicts, got: %v", err)
	}
}

func TestRunMerge_HappyPath_RemovesWorktreeAndRunsValidation(t *testing.T) {
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
	_ = os.WriteFile(filepath.Join(d, ".gitignore"), []byte(".worktrees/\n"), 0644)
	// Minimal centinela.toml so runValidateForMerge can load config.
	// validate.commands is empty — gates run but no shell commands.
	_ = os.WriteFile(filepath.Join(d, "centinela.toml"),
		[]byte("[validate]\ncommands = []\n[gates]\nfile_size = false\n"), 0644)
	for _, args := range [][]string{{"add", "."}, {"commit", "-q", "-m", "seed"}} {
		c := exec.Command("git", args...)
		c.Dir = d
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	os.Chdir(d) //nolint:errcheck

	// Provision and commit inside an isolated worktree, then merge.
	wt := filepath.Join(d, ".worktrees", "omega")
	c := exec.Command("git", "worktree", "add", ".worktrees/omega", "-b", "omega")
	c.Dir = d
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add: %v\n%s", err, out)
	}
	_ = os.WriteFile(filepath.Join(wt, "feature.txt"), []byte("omega\n"), 0644)
	for _, args := range [][]string{{"add", "."}, {"commit", "-q", "-m", "omega commit"}} {
		c := exec.Command("git", args...)
		c.Dir = wt
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v in wt: %v\n%s", args, err, out)
		}
	}

	if err := runMerge(nil, []string{"omega"}); err != nil {
		t.Fatalf("runMerge happy path: %v", err)
	}
	if _, err := os.Stat(wt); !os.IsNotExist(err) {
		t.Fatalf("worktree should be removed after clean merge; err=%v", err)
	}
}
