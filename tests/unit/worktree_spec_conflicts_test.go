package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// writeSpec creates `<root>/specs/<name>.feature` with the given body.
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

// makeWorktreeDir creates `.worktrees/<feat>/specs/<name>.feature`.
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

func TestDetectSpecConflicts_SameGivenSameThen_NoFlag(t *testing.T) {
	repo := t.TempDir()
	writeSpec(t, repo, "alpha", `Feature: A
  Scenario: same outcome
    Given identical context
    Then identical result
`)
	makeWorktreeSpec(t, repo, "beta", "beta", `Feature: B
  Scenario: also same
    Given identical context
    Then identical result
`)
	conflicts := worktree.DetectSpecConflicts(repo, "alpha")
	if len(conflicts) != 0 {
		t.Fatalf("agreeing scenarios should not conflict, got %v", conflicts)
	}
}

func TestDetectSpecConflicts_NoSpecsDirectory_NoError(t *testing.T) {
	repo := t.TempDir()
	conflicts := worktree.DetectSpecConflicts(repo, "ghost")
	if len(conflicts) != 0 {
		t.Fatalf("no specs should yield no conflicts, got %v", conflicts)
	}
}

func TestDetectSpecConflicts_SameOwnerIsNotConflict(t *testing.T) {
	repo := t.TempDir()
	// Two scenarios in the same file (same owner) with conflicting Then must
	// not produce a conflict — the detector ignores intra-feature collisions.
	writeSpec(t, repo, "solo", `Feature: S
  Scenario: a
    Given a context
    Then result is A
  Scenario: b
    Given a context
    Then result is B
`)
	conflicts := worktree.DetectSpecConflicts(repo, "solo")
	if len(conflicts) != 0 {
		t.Fatalf("intra-feature conflicts should be ignored, got %v", conflicts)
	}
}

func TestFormatSpecConflicts_Empty(t *testing.T) {
	if got := worktree.FormatSpecConflicts(nil); got != "" {
		t.Fatalf("expected empty string for empty conflict slice, got %q", got)
	}
}
