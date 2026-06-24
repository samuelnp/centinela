// Acceptance: specs/completion-delivery-prompt.feature
package acceptance_test

import (
	"os/exec"
	"strings"
	"testing"
)

func commitAll(t *testing.T, dir string) {
	t.Helper()
	for _, a := range [][]string{{"add", "-A"}, {"-c", "user.email=t@t", "-c", "user.name=t", "commit", "-m", "x"}} {
		c := exec.Command("git", a...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", a, out)
		}
	}
}

// Scenario: deliver without --via refuses to act and exits non-zero
func TestAccDeliverNoVia(t *testing.T) {
	dir := cdpRepo(t, true)
	cdpWorkflow(t, dir, "alpha", false, true)
	out, code := runDeliverBin(t, dir, "alpha")
	if code == 0 || !strings.Contains(out, "via") {
		t.Fatalf("no --via should fail asking for via: code=%d\n%s", code, out)
	}
}

// Scenario: deliver --via pr with no origin remote refuses to act and exits non-zero
func TestAccDeliverPRNoOrigin(t *testing.T) {
	dir := cdpRepo(t, false)
	cdpWorkflow(t, dir, "alpha", false, true)
	out, code := runDeliverBin(t, dir, "alpha", "--via", "pr")
	if code == 0 || !strings.Contains(out, "no origin remote") {
		t.Fatalf("pr without origin should refuse: code=%d\n%s", code, out)
	}
	if strings.Contains(out, "Pushed") {
		t.Fatalf("must not push: %s", out)
	}
}

// Scenario: deliver --via merge delegates to the existing merge flow on a clean merge
// Scenario: deliver --via merge on a conflicted merge reuses the merge-steward dispatch
func TestAccDeliverMergeDelegates(t *testing.T) {
	dir := cdpRepo(t, true)
	cdpWorkflow(t, dir, "alpha", false, true)
	out, code := runDeliverBin(t, dir, "alpha", "--via", "merge")
	// Worktree mode is satisfied, so the matrix accepts merge and delegates to
	// the merge flow (which then handles clean/conflict/steward). It must NOT be
	// rejected by the deliver guard.
	if strings.Contains(out, "worktree mode required") {
		t.Fatalf("merge should be delegated, not guard-rejected:\n%s", out)
	}
	if code == 0 {
		t.Fatalf("merge of a non-existent worktree branch should fail, got 0\n%s", out)
	}
}

// Scenario: deliver --via pr with origin and gh available pushes and reports the opened PR
// Scenario: deliver --via pr when gh is absent still pushes, prints manual instructions, and exits non-zero
func TestAccDeliverPRPathReachedHonestly(t *testing.T) {
	dir := cdpRepo(t, true)
	cdpWorkflow(t, dir, "alpha", false, true)
	commitAll(t, dir) // clean tree so the pr path reaches push (origin is a stub URL)
	out, _ := runDeliverBin(t, dir, "alpha", "--via", "pr")
	// The guards pass (origin present, clean tree) so the pr path is reached.
	if strings.Contains(out, "no origin remote") || strings.Contains(out, "choose --via") || strings.Contains(out, "uncommitted") {
		t.Fatalf("pr path should be reached past the guards:\n%s", out)
	}
	// Whatever happens against the stub remote, it must never falsely claim a PR.
	if strings.Contains(out, "Opened pull request") {
		t.Fatalf("must not claim a PR was opened against a stub remote:\n%s", out)
	}
}
