package workflow

import "github.com/samuelnp/centinela/internal/acceptance"

// hasAcceptanceExecutionCommand answers "is any of these an acceptance
// execution". The per-command predicate moved to the internal/acceptance leaf
// so cmd/ and internal/verify can share it without importing this package; the
// classification itself is unchanged, and this file's existing tests are the
// proof of that.
func hasAcceptanceExecutionCommand(commands []string) bool {
	return acceptance.AnyExecutionCommand(commands)
}
