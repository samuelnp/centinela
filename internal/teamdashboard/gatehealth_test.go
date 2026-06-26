package teamdashboard

import (
	"reflect"
	"testing"

	"github.com/samuelnp/centinela/internal/insights"
	"github.com/samuelnp/centinela/internal/telemetry"
)

func gateEvent(gate string) telemetry.Event {
	return telemetry.Event{Type: telemetry.TypeGateFailure, Gate: gate}
}

func TestGatehealth_RanksAndMatchesInsights(t *testing.T) {
	events := []telemetry.Event{
		gateEvent("coverage"), gateEvent("coverage"), gateEvent("coverage"),
		gateEvent("import-graph"), gateEvent("import-graph"),
		{Type: telemetry.TypeStepAdvanced, Gate: "coverage"}, // excluded
		{Type: telemetry.TypeBlock, Gate: "import-graph"},    // excluded
	}
	got := gatehealth(events, gateTopN)
	if len(got) != 2 {
		t.Fatalf("want 2 gates, got %d: %+v", len(got), got)
	}
	if got[0].Gate != "coverage" || got[0].Fails != 3 {
		t.Fatalf("top gate: %+v", got[0])
	}
	if got[1].Gate != "import-graph" || got[1].Fails != 2 {
		t.Fatalf("second gate: %+v", got[1])
	}
	// Must mirror insights.Gates verbatim.
	want := insights.Gates(events, gateTopN)
	mapped := make([]GateHealth, 0, len(want))
	for _, c := range want {
		mapped = append(mapped, GateHealth{Gate: c.Key, Fails: c.Count})
	}
	if !reflect.DeepEqual(got, mapped) {
		t.Fatalf("gatehealth diverged from insights.Gates: %+v vs %+v", got, mapped)
	}
}

func TestGatehealth_EmptyAndNoGateFailures(t *testing.T) {
	if g := gatehealth(nil, gateTopN); len(g) != 0 {
		t.Fatalf("nil events => empty, got %+v", g)
	}
	only := []telemetry.Event{{Type: telemetry.TypeBlock}, {Type: telemetry.TypeStepAdvanced}}
	if g := gatehealth(only, gateTopN); len(g) != 0 {
		t.Fatalf("no gate-failure => empty, got %+v", g)
	}
}

func TestGatehealth_EmptyGateBucketsNone(t *testing.T) {
	g := gatehealth([]telemetry.Event{gateEvent("")}, gateTopN)
	if len(g) != 1 || g[0].Gate != "<none>" {
		t.Fatalf("empty gate must bucket under <none>, got %+v", g)
	}
}
