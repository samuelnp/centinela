package calibration

import (
	"testing"

	"github.com/samuelnp/centinela/internal/telemetry"
)

// stamped builds an event with a timestamp, for span coverage.
func stamped(model, typ, ts string) telemetry.Event {
	return telemetry.Event{Type: typ, Model: model, Timestamp: ts}
}

// findRec returns the ModelRecord for id, or fails.
func findRec(t *testing.T, r Report, id string) ModelRecord {
	t.Helper()
	for _, m := range r.Models {
		if m.Model == id {
			return m
		}
	}
	t.Fatalf("no record for %q in %+v", id, r.Models)
	return ModelRecord{}
}

// TestCalibrateEmpty — empty slice yields an empty-state Report.
func TestCalibrateEmpty(t *testing.T) {
	r := Calibrate(nil, nil)
	if r.ModelCount != 0 || len(r.Models) != 0 || r.SpanStart != "" || r.SpanEnd != "" {
		t.Fatalf("empty report wrong: %+v", r)
	}
}

// TestCalibrateEndToEnd — mixed models each get an independent classification,
// span is computed, and unattributed is bucketed.
func TestCalibrateEndToEnd(t *testing.T) {
	evs := []telemetry.Event{
		stamped("claude-opus-4-7", telemetry.TypeStepAdvanced, "2026-01-01T00:00:00Z"),
		stamped("claude-opus-4-7", telemetry.TypeStepAdvanced, "2026-01-02T00:00:00Z"),
		stamped("claude-opus-4-7", telemetry.TypeStepAdvanced, "2026-01-03T00:00:00Z"),
		stamped("claude-opus-4-7", telemetry.TypeStepAdvanced, "2026-01-04T00:00:00Z"),
		ev("claude-opus-4-7", telemetry.TypeGateFailure),
		ev("claude-sonnet-4-6", telemetry.TypeStepAdvanced), ev("claude-sonnet-4-6", telemetry.TypeStepAdvanced),
		ev("claude-sonnet-4-6", telemetry.TypeStepAdvanced),
		ev("claude-sonnet-4-6", telemetry.TypeGateFailure), ev("claude-sonnet-4-6", telemetry.TypeGateFailure),
		ev("claude-sonnet-4-6", telemetry.TypeGateFailure),
		stamped("", telemetry.TypeStepAdvanced, "2026-06-01T00:00:00Z"),
	}
	r := Calibrate(evs, nil)
	if r.ModelCount != 3 {
		t.Fatalf("ModelCount = %d, want 3", r.ModelCount)
	}
	if r.SpanStart != "2026-01-01T00:00:00Z" || r.SpanEnd != "2026-06-01T00:00:00Z" {
		t.Fatalf("span wrong: %q..%q", r.SpanStart, r.SpanEnd)
	}
	if findRec(t, r, "claude-opus-4-7").Verdict != WellCalibrated {
		t.Fatal("opus should be WellCalibrated")
	}
	if findRec(t, r, "claude-sonnet-4-6").Verdict != Undergoverned {
		t.Fatal("sonnet should be Undergoverned")
	}
	if findRec(t, r, unattributed).Verdict != Unclassified {
		t.Fatal("unattributed should be Unclassified")
	}
}

// TestCalibrateDeterministicSort — model id ascending with unattributed last.
func TestCalibrateDeterministicSort(t *testing.T) {
	evs := []telemetry.Event{
		ev("zeta-model", telemetry.TypeStepAdvanced),
		ev("alpha-model", telemetry.TypeStepAdvanced),
		ev("", telemetry.TypeStepAdvanced),
		ev("claude-haiku-4-5", telemetry.TypeStepAdvanced),
	}
	r := Calibrate(evs, nil)
	got := []string{r.Models[0].Model, r.Models[1].Model, r.Models[2].Model, r.Models[3].Model}
	want := []string{"alpha-model", "claude-haiku-4-5", "zeta-model", unattributed}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sort order = %v, want %v", got, want)
		}
	}
}
