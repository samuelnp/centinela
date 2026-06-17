package insights

import "github.com/samuelnp/centinela/internal/telemetry"

// stepsToGreen computes the mean attempts-to-green across the repo:
// (complete-rejected + step-advanced) / step-advanced. "Green" is a
// step-advanced event; each advance took 1 + its rejections attempts, so the
// ratio is total complete-attempts per successful advance. When there are zero
// advances the denominator is 0: HasValue is false and Mean stays 0 (no
// division, no panic) — the renderer prints "n/a".
func stepsToGreen(events []telemetry.Event) StepsStat {
	var advances, rejections int
	for _, e := range events {
		switch e.Type {
		case telemetry.TypeStepAdvanced:
			advances++
		case telemetry.TypeCompleteRejected:
			rejections++
		}
	}
	s := StepsStat{Advances: advances, Rejections: rejections}
	if advances > 0 {
		s.Mean = float64(rejections+advances) / float64(advances)
		s.HasValue = true
	}
	return s
}

// span returns the earliest and latest event timestamps. Timestamps are
// RFC3339 UTC, which sorts lexically, so a string min/max is correct. Empty
// timestamps are ignored; an event-free (or timestamp-free) log yields ("", "").
func span(events []telemetry.Event) (start, end string) {
	for _, e := range events {
		if e.Timestamp == "" {
			continue
		}
		if start == "" || e.Timestamp < start {
			start = e.Timestamp
		}
		if end == "" || e.Timestamp > end {
			end = e.Timestamp
		}
	}
	return start, end
}
