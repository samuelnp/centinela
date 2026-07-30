package worktree_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// writeSpec writes a spec into the main checkout.
func writeSpec(t *testing.T, root, name, body string) {
	t.Helper()
	dir := filepath.Join(root, "specs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir specs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".feature"), []byte(body), 0644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
}

// makeWorktreeSpec writes a spec into one feature worktree.
func makeWorktreeSpec(t *testing.T, repo, feat, name, body string) {
	t.Helper()
	dir := filepath.Join(repo, ".worktrees", feat, "specs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".feature"), []byte(body), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// TRUE POSITIVE: two in-flight worktrees carry the same (file, scenario) with
// different outcomes — merging one silently overwrites the other.
func TestDetectSpecConflicts_TwoWorktreesDivergentThen_Flags(t *testing.T) {
	repo := t.TempDir()
	makeWorktreeSpec(t, repo, "zeta", "login", `Feature: Login
  Scenario: shared
    Given the same context
    Then result is A
`)
	makeWorktreeSpec(t, repo, "eta", "login", `Feature: Login
  Scenario: shared
    Given the same context
    Then result is B
`)
	conflicts := worktree.DetectSpecConflicts(repo, "zeta")
	if len(conflicts) != 1 {
		t.Fatalf("expected exactly one conflict, got %d: %v", len(conflicts), conflicts)
	}
	got := worktree.FormatSpecConflicts(conflicts)
	for _, want := range []string{"the same context", "shared", "zeta", "eta", "login.feature"} {
		if !strings.Contains(got, want) {
			t.Fatalf("formatted conflict missing %q: %q", want, got)
		}
	}
}

func TestDetectSpecConflicts_NoSpecsDirectory_NoError(t *testing.T) {
	repo := t.TempDir()
	conflicts := worktree.DetectSpecConflicts(repo, "ghost")
	if len(conflicts) != 0 {
		t.Fatalf("no specs should yield no conflicts, got %v", conflicts)
	}
}

func TestFormatSpecConflicts_Empty(t *testing.T) {
	if got := worktree.FormatSpecConflicts(nil); got != "" {
		t.Fatalf("expected empty string for empty conflict slice, got %q", got)
	}
}
