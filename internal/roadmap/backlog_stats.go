package roadmap

// BacklogStats is the footer arithmetic: how much debt there is, how much of
// it has rotted, and what the oldest item is. OldestKnown is false when no
// finding carries a parseable deferredAt.
type BacklogStats struct {
	Total         int
	Stale         int
	ThresholdDays int
	OldestDays    int
	OldestSlug    string
	OldestKnown   bool
}

// SummarizeBacklog reduces an aged list (as returned by AgeBacklog, so already
// oldest-first) to its counts. thresholdDays is carried through for rendering
// so the caller cannot report a count against a different threshold.
func SummarizeBacklog(aged []Aged, thresholdDays int) BacklogStats {
	if thresholdDays <= 0 {
		thresholdDays = DefaultStaleDays
	}
	stats := BacklogStats{Total: len(aged), ThresholdDays: thresholdDays}
	for _, a := range aged {
		if a.Stale {
			stats.Stale++
		}
		if a.KnownAge && (!stats.OldestKnown || a.AgeDays > stats.OldestDays) {
			stats.OldestKnown, stats.OldestDays, stats.OldestSlug = true, a.AgeDays, a.Name
		}
	}
	return stats
}
