package insights

import (
	"reflect"
	"testing"

	"github.com/samuelnp/centinela/internal/telemetry"
)

// mixed builds one event of each type with timestamps for span coverage.
func mixed() []telemetry.Event {
	return []telemetry.Event{
		{Type: telemetry.TypeBlock, Reason: "out-of-step", FileType: "plan", Timestamp: "2026-01-01T00:00:00Z"},
		{Type: telemetry.TypeGateFailure, Gate: "coverage", Feature: "alpha", Timestamp: "2026-02-01T00:00:00Z"},
		{Type: telemetry.TypeVerifyRejection, Feature: "alpha", Timestamp: "2026-03-01T00:00:00Z"},
		{Type: telemetry.TypeCompleteRejected, Feature: "beta", Timestamp: "2026-04-01T00:00:00Z"},
		{Type: telemetry.TypeStepAdvanced, Feature: "alpha", Timestamp: "2026-06-01T12:00:00Z"},
	}
}

// Compute fills every section end-to-end with correct counts and span.
func TestComputeEndToEnd(t *testing.T) {
	r := Compute(mixed(), 5)
	if r.EventCount != 5 {
		t.Fatalf("EventCount = %d, want 5", r.EventCount)
	}
	if r.SpanStart != "2026-01-01T00:00:00Z" || r.SpanEnd != "2026-06-01T12:00:00Z" {
		t.Fatalf("span = %q..%q", r.SpanStart, r.SpanEnd)
	}
	if len(r.Blocks) != 1 || r.Blocks[0].Key != "out-of-step · plan" {
		t.Fatalf("blocks = %+v", r.Blocks)
	}
	if len(r.Gates) != 1 || r.Gates[0].Key != "coverage" {
		t.Fatalf("gates = %+v", r.Gates)
	}
	// alpha: gate-failure + verify-rejection = 2; beta: complete-rejected = 1.
	if len(r.Rework) != 2 || r.Rework[0].Key != "alpha" || r.Rework[0].Count != 2 {
		t.Fatalf("rework = %+v", r.Rework)
	}
	// 1 advance, 1 rejection ⇒ 2.00.
	if !r.StepsToGreen.HasValue || r.StepsToGreen.Mean != 2.0 {
		t.Fatalf("steps = %+v", r.StepsToGreen)
	}
}

// Empty slice ⇒ empty-state Report, no panic.
func TestComputeEmpty(t *testing.T) {
	r := Compute(nil, 5)
	if r.EventCount != 0 || r.SpanStart != "" || len(r.Blocks) != 0 ||
		len(r.Gates) != 0 || len(r.Rework) != 0 || r.StepsToGreen.HasValue {
		t.Fatalf("non-empty empty-state report: %+v", r)
	}
}

// Compute is deterministic: same input yields a deeply-equal Report.
func TestComputeDeterministic(t *testing.T) {
	a := Compute(mixed(), 5)
	b := Compute(mixed(), 5)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("non-deterministic: %+v vs %+v", a, b)
	}
}

// topN bounds each ranked section in Compute.
func TestComputeTopN(t *testing.T) {
	ev := []telemetry.Event{
		{Type: telemetry.TypeBlock, Reason: "a"}, {Type: telemetry.TypeBlock, Reason: "b"},
		{Type: telemetry.TypeBlock, Reason: "c"},
	}
	if r := Compute(ev, 2); len(r.Blocks) != 2 {
		t.Fatalf("blocks top 2 = %d", len(r.Blocks))
	}
}
