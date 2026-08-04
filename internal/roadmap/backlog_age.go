package roadmap

import (
	"sort"
	"time"
)

// DefaultStaleDays is one delivery cadence. A finding older than this has
// survived a full cycle without being scheduled.
const DefaultStaleDays = 14

// Aged pairs a Backlog finding with its computed age. KnownAge is false when
// deferredAt is missing or unparseable — a hand-edited or legacy entry. Those
// count as Stale and sort first, deliberately the fail-loud direction: an entry
// with no clock surfaces instead of hiding at the bottom of the list.
type Aged struct {
	Feature
	AgeDays  int
	KnownAge bool
	Stale    bool
}

// AgeBacklog returns every Backlog finding with its age, oldest first —
// unknown ages before all known ones, ties broken by name so the order is
// deterministic. thresholdDays <= 0 falls back to DefaultStaleDays.
//
// "Stale" is strictly older than the threshold: a finding deferred exactly N
// days ago is not yet older than N days.
func AgeBacklog(r *Roadmap, now time.Time, thresholdDays int) []Aged {
	if thresholdDays <= 0 {
		thresholdDays = DefaultStaleDays
	}
	findings := BacklogFeatures(r)
	out := make([]Aged, 0, len(findings))
	for _, f := range findings {
		a := Aged{Feature: f}
		if t, err := time.Parse(time.RFC3339, f.DeferredAt); err == nil {
			a.KnownAge = true
			a.AgeDays = int(now.Sub(t).Hours() / 24)
		}
		a.Stale = !a.KnownAge || a.AgeDays > thresholdDays
		out = append(out, a)
	}
	sort.SliceStable(out, func(i, j int) bool { return olderFirst(out[i], out[j]) })
	return out
}

// olderFirst orders unknown ages first, then descending age, then by name.
func olderFirst(a, b Aged) bool {
	if a.KnownAge != b.KnownAge {
		return !a.KnownAge
	}
	if a.KnownAge && a.AgeDays != b.AgeDays {
		return a.AgeDays > b.AgeDays
	}
	return a.Name < b.Name
}

// StaleOnly filters an aged list to the findings past the threshold, keeping
// the order AgeBacklog established.
func StaleOnly(aged []Aged) []Aged {
	out := make([]Aged, 0, len(aged))
	for _, a := range aged {
		if a.Stale {
			out = append(out, a)
		}
	}
	return out
}
