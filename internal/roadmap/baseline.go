package roadmap

import "strings"

// BaselinePhaseName is the canonical name of the schedule-exempt phase that
// records already-built capability discovered by `centinela roadmap brownfield`.
// Like the Backlog phase, its features are not schedulable work — they document
// what already exists so it is never re-planned — so they are excluded from the
// status counts, the validate coverage set, and readiness, via the shared
// isNonSchedulablePhase predicate.
const BaselinePhaseName = "Baseline"

// isBaselinePhaseName reports whether a phase name denotes the Baseline phase.
// Mirrors isBacklogPhaseName: lower-cased, trimmed, exact canonical match.
func isBaselinePhaseName(name string) bool {
	return strings.EqualFold(strings.TrimSpace(name), BaselinePhaseName)
}

// IsBaselinePhaseName is the exported form for callers outside this package
// (e.g. the UI render layer or the brownfield generator).
func IsBaselinePhaseName(name string) bool {
	return isBaselinePhaseName(name)
}

// isNonSchedulablePhase reports whether a phase's features are exempt from the
// status/coverage/readiness logic. It is the single conceptual place that decides
// "these entries are not schedulable work": both the Backlog phase (deferred
// findings) and the Baseline phase (already-built capability) qualify. Future
// schedule-exempt conventions add their predicate here in exactly one spot.
func isNonSchedulablePhase(name string) bool {
	return isBacklogPhaseName(name) || isBaselinePhaseName(name)
}
