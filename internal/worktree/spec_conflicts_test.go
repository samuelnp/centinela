package worktree_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

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

func TestDetectSpecConflicts_DifferentThenSameGiven_Flags(t *testing.T) {
	repo := t.TempDir()
	writeSpec(t, repo, "zeta", `Feature: Z
  Scenario: shared
    Given the same context
    Then result is A
`)
	makeWorktreeSpec(t, repo, "eta", "eta", `Feature: E
  Scenario: shared
    Given the same context
    Then result is B
`)
	conflicts := worktree.DetectSpecConflicts(repo, "zeta")
	if len(conflicts) == 0 {
		t.Fatal("expected conflict, got none")
	}
	got := worktree.FormatSpecConflicts(conflicts)
	if !strings.Contains(got, "the same context") {
		t.Fatalf("formatter missing Given context: %q", got)
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
