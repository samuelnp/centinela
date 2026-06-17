package insights

import (
	"testing"

	"github.com/samuelnp/centinela/internal/telemetry"
)

func rw(typ, feature string) telemetry.Event {
	return telemetry.Event{Type: typ, Feature: feature}
}

// rework score = gate-failure + verify-rejection + complete-rejected per feature.
func TestReworkSumsFrictionTypes(t *testing.T) {
	ev := []telemetry.Event{
		rw(telemetry.TypeGateFailure, "alpha"), rw(telemetry.TypeGateFailure, "alpha"),
		rw(telemetry.TypeVerifyRejection, "alpha"),
		rw(telemetry.TypeCompleteRejected, "beta"), rw(telemetry.TypeGateFailure, "beta"),
	}
	got := rework(ev, 5)
	if len(got) != 2 || got[0].Key != "alpha" || got[0].Count != 3 {
		t.Fatalf("first = %+v (all %+v)", got[0], got)
	}
	if got[1].Key != "beta" || got[1].Count != 2 {
		t.Fatalf("second = %+v", got[1])
	}
}

// Events with an empty Feature are excluded (no anonymous bucket).
func TestReworkExcludesEmptyFeature(t *testing.T) {
	ev := []telemetry.Event{rw(telemetry.TypeGateFailure, ""), rw(telemetry.TypeGateFailure, "alpha")}
	got := rework(ev, 5)
	if len(got) != 1 || got[0].Key != "alpha" || got[0].Count != 1 {
		t.Fatalf("got = %+v", got)
	}
}

// step-advanced is not friction and must not be counted.
func TestReworkExcludesStepAdvanced(t *testing.T) {
	ev := []telemetry.Event{
		rw(telemetry.TypeStepAdvanced, "alpha"), rw(telemetry.TypeStepAdvanced, "alpha"),
		rw(telemetry.TypeGateFailure, "alpha"),
	}
	got := rework(ev, 5)
	if len(got) != 1 || got[0].Count != 1 {
		t.Fatalf("got = %+v", got)
	}
}

// Equal counts break ties by feature name ascending.
func TestReworkTieBreakFeatureAsc(t *testing.T) {
	got := rework([]telemetry.Event{rw(telemetry.TypeGateFailure, "zeta"), rw(telemetry.TypeGateFailure, "alpha")}, 5)
	if got[0].Key != "alpha" || got[1].Key != "zeta" {
		t.Fatalf("order = %+v", got)
	}
}

// topN truncates the rework section.
func TestReworkTopNTruncates(t *testing.T) {
	ev := []telemetry.Event{rw(telemetry.TypeGateFailure, "a"), rw(telemetry.TypeGateFailure, "b"), rw(telemetry.TypeGateFailure, "c")}
	if got := rework(ev, 2); len(got) != 2 {
		t.Fatalf("top 2 = %d entries", len(got))
	}
}
