// Acceptance: specs/workflow-revise-loop.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// rlReject seeds a canonical workflow at current, runs revise with args, and
// asserts the command is rejected, the output contains want, and the persisted
// current step is unchanged (no state mutation).
func rlReject(t *testing.T, current, want string, args ...string) {
	t.Helper()
	dir := rlDir(t)
	rlState(t, dir, "my-feature", rlCanonical, current)
	out, code := runCent(t, buildCent(t), dir,
		append([]string{"revise", "my-feature"}, args...)...)
	if code == 0 {
		t.Fatalf("want non-zero exit, got 0: %s", out)
	}
	if want != "" && !strings.Contains(out, want) {
		t.Fatalf("output %q must contain %q", out, want)
	}
	if got := rlLoad(t, dir, "my-feature")["currentStep"]; got != current {
		t.Fatalf("state mutated: current = %v, want %s", got, current)
	}
}

// Scenario: Negative — revise without --reason is rejected
func TestRL_MissingReason(t *testing.T) {
	rlReject(t, "validate", "reason", "--to", "code")
}

// Scenario: Negative — empty or whitespace-only --reason is rejected
func TestRL_WhitespaceReason(t *testing.T) {
	rlReject(t, "validate", "empty", "--to", "code", "--reason", "   ")
}

// Scenario: Negative — revise to a forward step is rejected
func TestRL_ForwardTarget(t *testing.T) {
	rlReject(t, "code", "strictly before", "--to", "tests", "--reason", "jump forward")
}

// Scenario: Negative — revise to the current step is rejected
func TestRL_EqualTarget(t *testing.T) {
	rlReject(t, "validate", "strictly before", "--to", "validate", "--reason", "same step")
}

// Scenario: Negative — revise to an unknown step name is rejected
func TestRL_UnknownStep(t *testing.T) {
	rlReject(t, "validate", "deploy", "--to", "deploy", "--reason", "unknown")
}

// Scenario: Negative — revising a done workflow is rejected
func TestRL_DoneWorkflow(t *testing.T) {
	rlReject(t, "done", "completed workflow", "--to", "validate", "--reason", "reopen")
}
