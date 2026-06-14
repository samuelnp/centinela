package insights

import "github.com/samuelnp/centinela/internal/telemetry"

// reworkType reports whether an event type counts toward a feature's rework
// score: governance friction before green (gate-failure + verify-rejection +
// complete-rejected). step-advanced is explicitly excluded.
func reworkType(t string) bool {
	switch t {
	case telemetry.TypeGateFailure,
		telemetry.TypeVerifyRejection,
		telemetry.TypeCompleteRejected:
		return true
	default:
		return false
	}
}

// rework ranks features by rework score, count desc then feature asc, truncated
// to topN. Events with an empty Feature are excluded (no anonymous bucket).
func rework(events []telemetry.Event, topN int) []Count {
	m := make(map[string]int)
	for _, e := range events {
		if e.Feature == "" || !reworkType(e.Type) {
			continue
		}
		m[e.Feature]++
	}
	return rankTop(m, topN)
}
