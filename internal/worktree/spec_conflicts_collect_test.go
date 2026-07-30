package worktree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// Non-.feature files and nested directories inside specs/ are skipped.
func TestDetectSpecConflicts_IgnoresNonFeatureEntries(t *testing.T) {
	repo := t.TempDir()
	makeWorktreeSpec(t, repo, "zeta", "d", "Feature: D\n  Scenario: s\n    Given ctx\n    Then A\n")
	makeWorktreeSpec(t, repo, "eta", "d", "Feature: D\n  Scenario: s\n    Given ctx\n    Then B\n")
	specs := filepath.Join(repo, ".worktrees", "eta", "specs")
	if err := os.WriteFile(filepath.Join(specs, "notes.md"), []byte("noise"), 0644); err != nil {
		t.Fatalf("write notes: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(specs, "nested.feature"), 0755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	if got := worktree.DetectSpecConflicts(repo, "zeta"); len(got) != 1 {
		t.Fatalf("expected 1 conflict, got %d: %v", len(got), got)
	}
}

// An unreadable spec file is skipped rather than aborting detection.
func TestDetectSpecConflicts_UnreadableSpecSkipped(t *testing.T) {
	repo := t.TempDir()
	makeWorktreeSpec(t, repo, "zeta", "d", "Feature: D\n  Scenario: s\n    Given ctx\n    Then A\n")
	makeWorktreeSpec(t, repo, "eta", "d", "Feature: D\n  Scenario: s\n    Given ctx\n    Then B\n")
	locked := filepath.Join(repo, ".worktrees", "eta", "specs", "locked.feature")
	if err := os.WriteFile(locked, []byte("Feature: L\n"), 0000); err != nil {
		t.Fatalf("write locked: %v", err)
	}
	if got := worktree.DetectSpecConflicts(repo, "zeta"); len(got) != 1 {
		t.Fatalf("unreadable spec should be skipped, got %d: %v", len(got), got)
	}
}

// A worktree that carries no specs at all contributes nothing.
func TestDetectSpecConflicts_WorktreeWithoutSpecsIgnored(t *testing.T) {
	repo := t.TempDir()
	makeWorktreeSpec(t, repo, "zeta", "d", "Feature: D\n  Scenario: s\n    Given ctx\n    Then A\n")
	if err := os.MkdirAll(filepath.Join(repo, ".worktrees", "empty"), 0755); err != nil {
		t.Fatalf("mkdir empty: %v", err)
	}
	if got := worktree.DetectSpecConflicts(repo, "zeta"); len(got) != 0 {
		t.Fatalf("spec-less worktree must contribute nothing, got %v", got)
	}
}

// Scenarios missing a Given or Then are incomplete and never compared.
func TestDetectSpecConflicts_IncompleteScenarioIgnored(t *testing.T) {
	repo := t.TempDir()
	makeWorktreeSpec(t, repo, "zeta", "d", "Feature: D\n  Scenario: s\n    Given ctx\n")
	makeWorktreeSpec(t, repo, "eta", "d", "Feature: D\n  Scenario: s\n    Given ctx\n    Then B\n")
	got := worktree.DetectSpecConflicts(repo, "zeta")
	if len(got) != 1 {
		t.Fatalf("a missing Then is itself a divergence, got %d: %v", len(got), got)
	}
}
