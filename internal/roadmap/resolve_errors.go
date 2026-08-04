package roadmap

import "fmt"

// PhaseConflictError reports a schedulable phase both sides changed. v1 refuses
// rather than guesses: silently unioning two real edits is how a finding gets
// lost, which is the failure this command exists to prevent.
type PhaseConflictError struct{ Phase string }

// Error names the phase the operator must resolve by hand.
func (e *PhaseConflictError) Error() string {
	return fmt.Sprintf("phase %q changed on both sides; resolve it by hand — "+
		"roadmap resolve only merges divergent Backlog findings", e.Phase)
}

// FindingConflictError reports a Backlog finding whose two sides cannot be
// reconciled by rule — one side edited it and the other deleted it, or both
// edited it differently. Neither presence nor deferredAt can decide those:
// dropping the entry discards a real edit, keeping either version discards the
// other. v1 refuses, exactly as it refuses a both-sides phase edit. Detail
// names which shape it was, so the message tells an operator what to look at.
type FindingConflictError struct {
	Slug   string
	Detail string
}

// Error names the finding the operator must resolve by hand.
func (e *FindingConflictError) Error() string {
	return fmt.Sprintf("Backlog finding %q %s; resolve it by hand — "+
		"roadmap resolve refuses to guess which version wins", e.Slug, e.Detail)
}

// The two irreconcilable shapes, phrased for the operator.
const (
	detailModifyDelete = "was modified on one side and deleted on the other"
	detailBothEdited   = "was modified differently on both sides"
)
