package worktree_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

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

func TestDetectSpecConflicts_SameOwnerIsNotConflict(t *testing.T) {
	repo := t.TempDir()
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

// TRUE POSITIVE: divergence in the Given clause of the same scenario counts
// too, not just the Then clause.
func TestDetectSpecConflicts_TwoWorktreesDivergentGiven_Flags(t *testing.T) {
	repo := t.TempDir()
	makeWorktreeSpec(t, repo, "zeta", "flow", `Feature: F
  Scenario: shared
    Given context one
    Then same result
`)
	makeWorktreeSpec(t, repo, "eta", "flow", `Feature: F
  Scenario: shared
    Given context two
    Then same result
`)
	if got := worktree.DetectSpecConflicts(repo, "zeta"); len(got) != 1 {
		t.Fatalf("divergent Given should conflict once, got %d: %v", len(got), got)
	}
}
