package insights

import "github.com/samuelnp/centinela/internal/telemetry"

// Compute is the single entry point: it turns a parsed event slice into a pure,
// deterministic Report. It is stdlib-only beyond the telemetry leaf and does no
// I/O, so it is trivially unit-testable. topN bounds each ranked section. An
// empty slice yields an empty-state Report (zero counts, empty span, no panic).
func Compute(events []telemetry.Event, topN int) Report {
	start, end := span(events)
	return Report{
		EventCount:   len(events),
		SpanStart:    start,
		SpanEnd:      end,
		Blocks:       blocks(events, topN),
		Gates:        gates(events, topN),
		Rework:       rework(events, topN),
		StepsToGreen: stepsToGreen(events),
	}
}
