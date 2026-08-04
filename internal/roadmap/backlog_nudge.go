package roadmap

import "time"

// Nudge carries the counts behind the completion prompt: deferral without a
// closure path is socially acceptable debt, so finishing the roadmap says how
// much of it is still open.
type Nudge struct {
	Total         int
	Stale         int
	ThresholdDays int
}

// NudgeFor reports whether the Backlog deserves a prompt, and the counts to
// print. It is true only when NO schedulable feature is still incomplete and
// the Backlog is non-empty — a pure predicate, so `complete` needs no roadmap
// reasoning of its own. A nil roadmap yields false.
func NudgeFor(r *Roadmap, now time.Time, thresholdDays int) (Nudge, bool) {
	if r == nil {
		return Nudge{}, false
	}
	if _, incomplete := FirstIncompleteSchedulable(r); incomplete {
		return Nudge{}, false
	}
	aged := AgeBacklog(r, now, thresholdDays)
	if len(aged) == 0 {
		return Nudge{}, false
	}
	stats := SummarizeBacklog(aged, thresholdDays)
	return Nudge{Total: stats.Total, Stale: stats.Stale, ThresholdDays: stats.ThresholdDays}, true
}
