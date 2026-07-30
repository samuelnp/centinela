package worktree_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// Regression suite for the spec-conflict-false-positives hotfix. Every case
// below made `centinela merge --via merge` unusable whenever any worktree
// existed; none of them is a real conflict.

const archetypesSpec = `Feature: Workflow archetypes
  Scenario: hotfix archetype resolves to code-tests-validate
    Given a feature started with the hotfix archetype
    Then the step list is code, tests, validate
  Scenario: active archetype is pinned
    Given a feature started with the hotfix archetype
    Then the archetype is recorded in the workflow state
`

// Class 1 + 2: byte-identical copies of one file in main and two worktrees,
// where the file legitimately has two companion scenarios sharing a Given.
func TestDetectSpecConflicts_IdenticalCopiesAndCompanionScenarios_NoFlag(t *testing.T) {
	repo := t.TempDir()
	writeSpec(t, repo, "workflow-archetypes", archetypesSpec)
	makeWorktreeSpec(t, repo, "token-diet", "workflow-archetypes", archetypesSpec)
	makeWorktreeSpec(t, repo, "other-feature", "workflow-archetypes", archetypesSpec)
	if got := worktree.DetectSpecConflicts(repo, "token-diet"); len(got) != 0 {
		t.Fatalf("identical copies / companion scenarios must not conflict, got %v", got)
	}
}

// Class 3: two different files on main share a Given clause. Pre-existing on
// main, unrelated to the merge — was flagged before this hotfix.
func TestDetectSpecConflicts_CrossFileSameGivenOnMain_NoFlag(t *testing.T) {
	repo := t.TempDir()
	writeSpec(t, repo, "actionable-evidence", `Feature: Actionable evidence
  Scenario: evidence is actionable
    Given an orchestrated step
    Then the evidence names the subagent
`)
	writeSpec(t, repo, "subagent-orchestration", `Feature: Orchestration
  Scenario: step delegates
    Given an orchestrated step
    Then the step is delegated to a subagent
`)
	makeWorktreeSpec(t, repo, "token-diet", "unrelated", `Feature: U
  Scenario: u
    Given something else
    Then something else happens
`)
	if got := worktree.DetectSpecConflicts(repo, "token-diet"); len(got) != 0 {
		t.Fatalf("cross-file same-Given on main must not conflict, got %v", got)
	}
}

// Class 4: main and the merging worktree disagree. That is supersession — the
// merge is precisely what applies it.
func TestDetectSpecConflicts_MainVersusMergingIsSupersession_NoFlag(t *testing.T) {
	repo := t.TempDir()
	writeSpec(t, repo, "login", `Feature: Login
  Scenario: routes after login
    Given user has account
    Then app routes to dashboard
`)
	makeWorktreeSpec(t, repo, "token-diet", "login", `Feature: Login
  Scenario: routes after login
    Given user has account
    Then app routes to onboarding
`)
	if got := worktree.DetectSpecConflicts(repo, "token-diet"); len(got) != 0 {
		t.Fatalf("main-vs-merging supersession must not conflict, got %v", got)
	}
}

// Inverted from TestDetectSpecConflicts_DifferentThenSameGiven_Flags, which
// asserted a false positive before the spec-conflict-false-positives hotfix:
// a main-checkout spec and a different worktree's differently-named spec are
// not comparable, and the merging feature has no worktree at all.
func TestDetectSpecConflicts_DifferentFilesDifferentOwners_NoFlag(t *testing.T) {
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
	if got := worktree.DetectSpecConflicts(repo, "zeta"); len(got) != 0 {
		t.Fatalf("different files in different owners must not conflict, got %v", got)
	}
}
