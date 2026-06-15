package calibration

import (
	"testing"

	"github.com/samuelnp/centinela/internal/telemetry"
)

// ev builds one event of the given type for the given model.
func ev(model, typ string) telemetry.Event {
	return telemetry.Event{Type: typ, Model: model}
}

// TestModelKeyFoldsEmpty — empty model folds to the "unattributed" bucket.
func TestModelKeyFoldsEmpty(t *testing.T) {
	if got := modelKey(""); got != unattributed {
		t.Fatalf("empty model key = %q, want %q", got, unattributed)
	}
	if got := modelKey("m"); got != "m" {
		t.Fatalf("model key = %q, want m", got)
	}
}

// TestReworkType — only gate/verify/complete-rejected count as rework.
func TestReworkType(t *testing.T) {
	rework := []string{telemetry.TypeGateFailure, telemetry.TypeVerifyRejection, telemetry.TypeCompleteRejected}
	for _, ty := range rework {
		if !reworkType(ty) {
			t.Fatalf("%q should be rework", ty)
		}
	}
	for _, ty := range []string{telemetry.TypeStepAdvanced, telemetry.TypeBlock, "other"} {
		if reworkType(ty) {
			t.Fatalf("%q should NOT be rework", ty)
		}
	}
}

// TestFrictionRateAndGuard — Rework sums all three rework types, Advances counts
// step-advanced, Rate = Rework/Advances with HasRate=true.
func TestFrictionRateAndGuard(t *testing.T) {
	evs := []telemetry.Event{
		ev("m", telemetry.TypeStepAdvanced), ev("m", telemetry.TypeStepAdvanced),
		ev("m", telemetry.TypeStepAdvanced), ev("m", telemetry.TypeStepAdvanced),
		ev("m", telemetry.TypeGateFailure), ev("m", telemetry.TypeVerifyRejection),
		ev("m", telemetry.TypeCompleteRejected), ev("m", telemetry.TypeBlock),
	}
	s := frictionByModel(evs)["m"]
	if s.Advances != 4 || s.Rework != 3 || s.GateFailures != 1 ||
		s.VerifyRejections != 1 || s.Blocks != 1 {
		t.Fatalf("counts wrong: %+v", s)
	}
	if !s.HasRate || s.Rate != 0.75 {
		t.Fatalf("rate = %v hasRate=%v, want 0.75 true", s.Rate, s.HasRate)
	}
}

// TestFrictionZeroAdvancesGuard — Advances==0 yields HasRate=false, Rate=0 (no
// division by zero / NaN) even with rework present.
func TestFrictionZeroAdvancesGuard(t *testing.T) {
	evs := []telemetry.Event{ev("m", telemetry.TypeGateFailure), ev("m", telemetry.TypeGateFailure)}
	s := frictionByModel(evs)["m"]
	if s.Advances != 0 || s.Rework != 2 {
		t.Fatalf("counts wrong: %+v", s)
	}
	if s.HasRate || s.Rate != 0 {
		t.Fatalf("zero-advance guard failed: rate=%v hasRate=%v", s.Rate, s.HasRate)
	}
}

// TestFrictionUnattributedBucket — empty-model events land in "unattributed".
func TestFrictionUnattributedBucket(t *testing.T) {
	out := frictionByModel([]telemetry.Event{ev("", telemetry.TypeStepAdvanced)})
	if _, ok := out[unattributed]; !ok {
		t.Fatalf("missing unattributed bucket: %+v", out)
	}
}
