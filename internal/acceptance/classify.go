// Package acceptance classifies validate commands as acceptance executions and
// parses runner reports for scenarios that did not actually assert.
//
// It imports the STANDARD LIBRARY ONLY. The classifier and the parser are
// needed by cmd/, internal/verify and internal/workflow alike; exporting them
// from internal/workflow would force an internal/verify → internal/workflow
// edge, so they live in a leaf that cannot participate in a cycle.
package acceptance

import "strings"

// IsExecutionCommand reports whether ONE command string is an acceptance
// execution. The body is byte-identical to the loop body of internal/workflow's
// hasAcceptanceExecutionCommand: this feature moves the predicate, it does not
// broaden it. Broadening the classification is explicitly out of scope, because
// the classifier gates the parse and the parse gates the verdict.
func IsExecutionCommand(cmd string) bool {
	c := strings.ToLower(strings.TrimSpace(cmd))
	if c == "" {
		return false
	}
	if strings.Contains(c, "tests/acceptance") {
		return true
	}
	if strings.Contains(c, "go test") && strings.Contains(c, "./...") {
		return true
	}
	if strings.Contains(c, "cucumber") || strings.Contains(c, "godog") || strings.Contains(c, "behave") {
		return true
	}
	if strings.Contains(c, "acceptance") && strings.Contains(c, "test") {
		return true
	}
	return false
}

// AnyExecutionCommand reports whether any command in the list is an acceptance
// execution. This is the any-of question internal/workflow's artifact gate asks,
// and the question the reused single-run outcome must be classified by: that
// outcome is labelled "validate.commands", not a real command string, so
// per-command classification cannot apply to it.
func AnyExecutionCommand(commands []string) bool {
	for _, cmd := range commands {
		if IsExecutionCommand(cmd) {
			return true
		}
	}
	return false
}
