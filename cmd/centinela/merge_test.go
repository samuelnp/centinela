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
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	specs := filepath.Join(d, "specs")
	_ = os.MkdirAll(specs, 0755)
	_ = os.WriteFile(filepath.Join(specs, "zeta.feature"),
		[]byte("Feature: Z\n  Scenario: clash\n    Given ctx\n    Then A\n"), 0644)

	other := filepath.Join(d, ".worktrees", "eta", "specs")
	_ = os.MkdirAll(other, 0755)
	_ = os.WriteFile(filepath.Join(other, "eta.feature"),
		[]byte("Feature: E\n  Scenario: clash\n    Given ctx\n    Then B\n"), 0644)

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
