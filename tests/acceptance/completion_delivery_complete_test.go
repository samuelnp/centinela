// Acceptance: specs/completion-delivery-prompt.feature
package acceptance_test

import (
	"strings"
	"testing"
)

func completeToDone(t *testing.T, withOrigin, worktree bool) (string, int) {
	t.Helper()
	dir := cdpRepo(t, withOrigin)
	cdpWorkflow(t, dir, "alpha", true, worktree)
	return runCent(t, buildCent(t), dir, "complete", "alpha")
}

// Scenario: Completion with an origin remote and worktree mode offers both PR and local merge
func TestAccCompleteOriginWorktreeBoth(t *testing.T) {
	out, code := completeToDone(t, true, true)
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "deliver alpha --via pr") || !strings.Contains(out, "deliver alpha --via merge") {
		t.Fatalf("expected both options:\n%s", out)
	}
}

// Scenario: Completion with no origin remote offers only the local-merge option
func TestAccCompleteNoOriginMergeOnly(t *testing.T) {
	out, code := completeToDone(t, false, true)
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "--via merge") || strings.Contains(out, "--via pr") {
		t.Fatalf("expected merge-only:\n%s", out)
	}
}

// Scenario: Completion in single-checkout mode with an origin remote offers only the PR option
func TestAccCompleteSingleCheckoutPROnly(t *testing.T) {
	out, code := completeToDone(t, true, false)
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "--via pr") || strings.Contains(out, "--via merge") {
		t.Fatalf("expected pr-only:\n%s", out)
	}
}

// Scenario: Completion with neither an origin remote nor worktree mode reports no delivery target
func TestAccCompleteNeitherNoTarget(t *testing.T) {
	out, code := completeToDone(t, false, false)
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "no delivery target") || strings.Contains(out, "--via") {
		t.Fatalf("expected no delivery target:\n%s", out)
	}
}

// Scenario: The completion directive never delivers by itself
func TestAccCompleteEmitsTextOnly(t *testing.T) {
	out, code := completeToDone(t, true, true)
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	if strings.Contains(out, "Pushed") || strings.Contains(out, "Merged") || strings.Contains(out, "Opened pull request") {
		t.Fatalf("completion must not deliver, only emit guidance:\n%s", out)
	}
	if !strings.Contains(out, "CENTINELA DIRECTIVE:") {
		t.Fatalf("expected the directive text:\n%s", out)
	}
}
