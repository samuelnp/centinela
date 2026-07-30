package worktree_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// A repeated scenario name must still be reported once, not once per copy —
// the unbounded report was 720KB of identical lines before this hotfix.
func TestDetectSpecConflicts_RepeatedScenarioReportedOnce(t *testing.T) {
	repo := t.TempDir()
	makeWorktreeSpec(t, repo, "zeta", "dup", `Feature: D
  Scenario: dup
    Given ctx
    Then A
  Scenario: dup
    Given ctx
    Then A
`)
	makeWorktreeSpec(t, repo, "eta", "dup", `Feature: D
  Scenario: dup
    Given ctx
    Then B
`)
	got := worktree.DetectSpecConflicts(repo, "zeta")
	if len(got) != 1 {
		t.Fatalf("repeated scenario must be reported once, got %d: %v", len(got), got)
	}
}

// Each other worktree is a distinct pairing and is reported separately.
func TestDetectSpecConflicts_TwoOtherWorktrees_OnePerPair(t *testing.T) {
	repo := t.TempDir()
	body := func(then string) string {
		return "Feature: D\n  Scenario: s\n    Given ctx\n    Then " + then + "\n"
	}
	makeWorktreeSpec(t, repo, "zeta", "d", body("A"))
	makeWorktreeSpec(t, repo, "eta", "d", body("B"))
	makeWorktreeSpec(t, repo, "theta", "d", body("C"))
	got := worktree.DetectSpecConflicts(repo, "zeta")
	if len(got) != 2 {
		t.Fatalf("expected one conflict per other worktree, got %d: %v", len(got), got)
	}
}

func buildSpec(t *testing.T, repo, feat, then string, n int) {
	t.Helper()
	var b strings.Builder
	b.WriteString("Feature: Many\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "  Scenario: s%d\n    Given ctx%d\n    Then %s\n", i, i, then)
	}
	makeWorktreeSpec(t, repo, feat, "many", b.String())
}

func TestFormatSpecConflicts_CapsLongReports(t *testing.T) {
	repo := t.TempDir()
	buildSpec(t, repo, "zeta", "A", 14)
	buildSpec(t, repo, "eta", "B", 14)
	conflicts := worktree.DetectSpecConflicts(repo, "zeta")
	if len(conflicts) != 14 {
		t.Fatalf("expected 14 conflicts, got %d", len(conflicts))
	}
	msg := worktree.FormatSpecConflicts(conflicts)
	if !strings.Contains(msg, "and 4 more") {
		t.Fatalf("formatter should cap and summarise the remainder: %q", msg)
	}
	if n := strings.Count(msg, "scenario "); n != 10 {
		t.Fatalf("formatter should print 10 entries, printed %d: %q", n, msg)
	}
}
