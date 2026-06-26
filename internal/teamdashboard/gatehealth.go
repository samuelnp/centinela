package teamdashboard

import (
	"github.com/samuelnp/centinela/internal/insights"
	"github.com/samuelnp/centinela/internal/telemetry"
)

// gatehealth tallies gate-failure events per gate by delegating to
// insights.Gates (an aggregator->aggregator edge, allowed) and mapping its
// ranked []insights.Count into []GateHealth. Reusing insights.Gates verbatim
// guarantees the board's ranking and counts never diverge from
// `centinela insights`. An empty Gate buckets under "<none>" (inherited from
// insights); no gate-failure events yields an empty slice (the empty state).
func gatehealth(events []telemetry.Event, topN int) []GateHealth {
	counts := insights.Gates(events, topN)
	out := make([]GateHealth, 0, len(counts))
	for _, c := range counts {
		out = append(out, GateHealth{Gate: c.Key, Fails: c.Count})
	}
	return out
}
