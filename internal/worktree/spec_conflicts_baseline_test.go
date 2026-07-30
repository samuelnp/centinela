package worktree_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func login(then string) string {
	return "Feature: Login\n  Scenario: routes\n    Given user has account\n    Then " + then + "\n"
}

// Every worktree is a full checkout, so a bystander carries main's untouched
// copy of the spec the merging feature edited. That is still supersession —
// nobody's work is overwritten. Regression for spec-conflict-false-positives.
func TestDetectSpecConflicts_BystanderCarriesMainBaseline_NoFlag(t *testing.T) {
	repo := t.TempDir()
	writeSpec(t, repo, "login", login("app routes to dashboard"))
	makeWorktreeSpec(t, repo, "token-diet", "login", login("app routes to onboarding"))
	makeWorktreeSpec(t, repo, "bystander", "login", login("app routes to dashboard"))
	if got := worktree.DetectSpecConflicts(repo, "token-diet"); len(got) != 0 {
		t.Fatalf("bystander holding main's copy must not conflict, got %v", got)
	}
}

// The mirror case: the merging feature never touched the spec, another
// worktree did. Merging cannot overwrite an edit it does not carry.
func TestDetectSpecConflicts_MergingCarriesMainBaseline_NoFlag(t *testing.T) {
	repo := t.TempDir()
	writeSpec(t, repo, "login", login("app routes to dashboard"))
	makeWorktreeSpec(t, repo, "token-diet", "login", login("app routes to dashboard"))
	makeWorktreeSpec(t, repo, "other", "login", login("app routes to onboarding"))
	if got := worktree.DetectSpecConflicts(repo, "token-diet"); len(got) != 0 {
		t.Fatalf("untouched merging copy must not conflict, got %v", got)
	}
}

// TRUE POSITIVE with a baseline present: both worktrees edited main's
// scenario in different directions.
func TestDetectSpecConflicts_BothEditedBaseline_Flags(t *testing.T) {
	repo := t.TempDir()
	writeSpec(t, repo, "login", login("app routes to dashboard"))
	makeWorktreeSpec(t, repo, "zeta", "login", login("app routes to dashboard v2"))
	makeWorktreeSpec(t, repo, "eta", "login", login("app routes to onboarding"))
	got := worktree.DetectSpecConflicts(repo, "zeta")
	if len(got) != 1 {
		t.Fatalf("two edits of the same baseline scenario must conflict once, got %d: %v", len(got), got)
	}
	if got[0].OwnerA != "zeta" || got[0].OwnerB != "eta" {
		t.Fatalf("conflict should name both worktrees, got %+v", got[0])
	}
}

// A scenario absent from main but introduced differently by two worktrees has
// no baseline to compare against and is still a conflict.
func TestDetectSpecConflicts_NoBaselineBothIntroduced_Flags(t *testing.T) {
	repo := t.TempDir()
	makeWorktreeSpec(t, repo, "zeta", "new", login("a"))
	makeWorktreeSpec(t, repo, "eta", "new", login("b"))
	if got := worktree.DetectSpecConflicts(repo, "zeta"); len(got) != 1 {
		t.Fatalf("independently introduced scenarios must conflict, got %d: %v", len(got), got)
	}
}
