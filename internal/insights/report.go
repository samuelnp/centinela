// Package insights computes read-only governance analytics over the telemetry
// event log. It is an aggregator over the telemetry leaf: it imports only
// internal/telemetry + stdlib, never cmd/ or internal/ui. Compute is pure and
// deterministic — every ranked section sorts a slice (count desc, key asc); no
// map is ranged in output order.
package insights

import "sort"

// Report is the pure, serializable analytics payload. Field names are a stable
// contract consumed by --json tooling; do not rename without bumping consumers.
type Report struct {
	EventCount   int       // total parsed events considered
	SpanStart    string    // earliest event timestamp (RFC3339), "" if none
	SpanEnd      string    // latest event timestamp, "" if none
	Blocks       []Count   // most-triggered blocks (ranked, top-N)
	Gates        []Count   // most-failed gates (ranked, top-N)
	Rework       []Count   // features by rework score (ranked, top-N)
	StepsToGreen StepsStat // mean attempts-to-green
}

// Count is a generic ranked bucket: a display key and its tally.
type Count struct {
	Key   string
	Count int
}

// StepsStat is the steps-to-green metric, computed without division by zero.
type StepsStat struct {
	Advances   int     // # step-advanced events (the denominator / "green"s)
	Rejections int     // # complete-rejected events
	Mean       float64 // (Rejections + Advances) / Advances; 0 when Advances==0
	HasValue   bool    // false when Advances==0 (renderer prints "n/a")
}

// rankTop converts a count map into a ranked slice (count desc, then key asc)
// and truncates to the top n. It never leaks map iteration order. A non-positive
// n yields an empty slice; n larger than the bucket count returns all buckets.
func rankTop(m map[string]int, n int) []Count {
	out := make([]Count, 0, len(m))
	for k, v := range m {
		out = append(out, Count{Key: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Key < out[j].Key
	})
	if n < 0 {
		n = 0
	}
	if n < len(out) {
		out = out[:n]
	}
	return out
}
