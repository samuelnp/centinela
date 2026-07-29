package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: a text conflict still dispatches the Merge Steward with no success claim
func TestAccMergeTextConflictKeepsWorktreeAndClaimsNothing(t *testing.T) {
	bin := buildCent(t)
	repo := mergeRepo(t, "high-score", true)
	wt := filepath.Join(repo, ".worktrees", "high-score")
	before := mtdaGit(t, repo, "rev-parse", "main")

	out, code := runCent(t, bin, repo, "merge", "high-score")
	if code == 0 {
		t.Fatalf("a text conflict must stop the merge (code=%d):\n%s", code, out)
	}
	if !strings.Contains(out, "CENTINELA DIRECTIVE:") ||
		!strings.Contains(out, "centinela merge --continue high-score") {
		t.Fatalf("the steward must be dispatched with a resume command:\n%s", out)
	}
	if strings.Contains(out, mtdcSuccess) || strings.Contains(out, "already merged") {
		t.Fatalf("a stalled merge must claim NOTHING:\n%s", out)
	}
	if mtdaGit(t, repo, "rev-parse", "main") != before {
		t.Fatalf("main must not be advanced by a conflicted merge:\n%s", out)
	}
	if _, e := os.Stat(wt); e != nil {
		t.Fatalf("the worktree must be kept for the steward: %v", e)
	}
	if m, e := os.Stat(filepath.Join(repo, ".workflow", "high-score-merge-pending.json")); e != nil || m.Size() == 0 {
		t.Fatalf("a pending marker must record the stall: %v", e)
	}
}

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: a validate failure after a clean text merge still dispatches the steward
func TestAccMergeValidateFailureDispatchesStewardWithoutClaim(t *testing.T) {
	bin := buildCent(t)
	repo := mtdfValidateFailRepo(t, "high-score") // merged tree fails validate
	wt := filepath.Join(repo, ".worktrees", "high-score")

	out, code := runCent(t, bin, wt, "merge", "high-score")
	if code == 0 {
		t.Fatalf("a validate failure must stop the merge (code=%d):\n%s", code, out)
	}
	if !strings.Contains(out, "CENTINELA DIRECTIVE:") {
		t.Fatalf("a validate failure must dispatch the steward:\n%s", out)
	}
	if strings.Contains(out, mtdcSuccess) || strings.Contains(out, "already merged") {
		t.Fatalf("no success may be claimed after a failed validate:\n%s", out)
	}
	if _, e := os.Stat(wt); e != nil {
		t.Fatalf("the worktree must be kept on a validate failure: %v", e)
	}
	marker := filepath.Join(repo, ".workflow", "high-score-merge-pending.json")
	data, e := os.ReadFile(marker)
	if e != nil {
		t.Fatalf("a pending marker must record the stall: %v", e)
	}
	if !strings.Contains(string(data), "post-merge-validate-failed") {
		t.Fatalf("the marker must name the validate-failure reason, got:\n%s", data)
	}
}
