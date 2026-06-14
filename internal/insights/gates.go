package insights

import "github.com/samuelnp/centinela/internal/telemetry"

// gates ranks gate-failure events by Gate, count desc then key asc, truncated to
// topN. An empty Gate field buckets under the <none> token.
func gates(events []telemetry.Event, topN int) []Count {
	m := make(map[string]int)
	for _, e := range events {
		if e.Type == telemetry.TypeGateFailure {
			m[orNone(e.Gate)]++
		}
	}
	return rankTop(m, topN)
}
