// Package acceptance classifies validate commands as acceptance executions and
// parses runner reports for scenarios that did not actually assert.
//
// It imports the STANDARD LIBRARY ONLY. The classifier and the parser are
// needed by cmd/, internal/verify and internal/workflow alike; exporting them
// from internal/workflow would force an internal/verify → internal/workflow
// edge, so they live in a leaf that cannot participate in a cycle.
package acceptance

// IsExecutionCommand reports whether ONE command string executes the acceptance
// tier. It is defined in terms of ScopeOf so the two can never drift, and the
// set of commands it accepts is byte-identical to the pre-move predicate in
// internal/workflow — whose own tests still pass unchanged. Broadening the
// classification remains out of scope.
//
// This answers "does the command RUN acceptance tests", which is the question
// internal/workflow's tests-step gate asks. Whether every skip the command
// reports is an ACCEPTANCE skip is a separate question — see Scope.
func IsExecutionCommand(cmd string) bool {
	return ScopeOf(cmd) != ScopeNone
}

// AnyExecutionCommand reports whether any command in the list is an acceptance
// execution. This is the any-of question internal/workflow's artifact gate asks,
// and the fallback classification for a reused outcome that carries no
// per-command judgement of its own.
func AnyExecutionCommand(commands []string) bool {
	for _, cmd := range commands {
		if IsExecutionCommand(cmd) {
			return true
		}
	}
	return false
}
