package insights

import (
	"testing"

	"github.com/samuelnp/centinela/internal/telemetry"
)

func gf(gate string) telemetry.Event {
	return telemetry.Event{Type: telemetry.TypeGateFailure, Gate: gate}
}

// gates buckets by Gate, count desc then key asc.
func TestGatesRanksByCountDesc(t *testing.T) {
	got := gates([]telemetry.Event{gf("coverage"), gf("coverage"), gf("import-graph")}, 5)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Key != "coverage" || got[0].Count != 2 {
		t.Fatalf("first = %+v", got[0])
	}
	if got[1].Key != "import-graph" || got[1].Count != 1 {
		t.Fatalf("second = %+v", got[1])
	}
}

// Equal counts break ties by gate name ascending.
func TestGatesTieBreakKeyAsc(t *testing.T) {
	got := gates([]telemetry.Event{gf("security"), gf("coverage")}, 5)
	if got[0].Key != "coverage" || got[1].Key != "security" {
		t.Fatalf("order = %+v", got)
	}
}

// An empty Gate buckets under the <none> token.
func TestGatesEmptyGateBucketsAsNone(t *testing.T) {
	got := gates([]telemetry.Event{gf("")}, 5)
	if len(got) != 1 || got[0].Key != "<none>" {
		t.Fatalf("key = %+v", got)
	}
}

// Non-gate-failure events are ignored.
func TestGatesIgnoresOtherTypes(t *testing.T) {
	ev := []telemetry.Event{{Type: telemetry.TypeBlock}, gf("coverage")}
	if got := gates(ev, 5); len(got) != 1 {
		t.Fatalf("got = %+v", got)
	}
}

// topN truncates the gates section.
func TestGatesTopNTruncates(t *testing.T) {
	ev := []telemetry.Event{gf("a"), gf("b"), gf("c")}
	if got := gates(ev, 2); len(got) != 2 {
		t.Fatalf("top 2 = %d entries", len(got))
	}
}
