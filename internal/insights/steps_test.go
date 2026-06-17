package insights

import (
	"testing"

	"github.com/samuelnp/centinela/internal/telemetry"
)

func adv() telemetry.Event { return telemetry.Event{Type: telemetry.TypeStepAdvanced} }
func rej() telemetry.Event { return telemetry.Event{Type: telemetry.TypeCompleteRejected} }

// Zero advances ⇒ HasValue false, Mean 0, no division by zero.
func TestStepsToGreenZeroAdvances(t *testing.T) {
	s := stepsToGreen([]telemetry.Event{rej(), rej()})
	if s.HasValue || s.Mean != 0 || s.Advances != 0 || s.Rejections != 2 {
		t.Fatalf("got %+v", s)
	}
}

// One advance, no rejections ⇒ mean 1.00.
func TestStepsToGreenSingleAdvance(t *testing.T) {
	s := stepsToGreen([]telemetry.Event{adv()})
	if !s.HasValue || s.Mean != 1.0 {
		t.Fatalf("got %+v", s)
	}
}

// One advance, one rejection ⇒ mean 2.00.
func TestStepsToGreenOneRejection(t *testing.T) {
	s := stepsToGreen([]telemetry.Event{adv(), rej()})
	if !s.HasValue || s.Mean != 2.0 {
		t.Fatalf("got %+v", s)
	}
}

// (4 advances + 2 rejections) / 4 advances = 1.50.
func TestStepsToGreenKnownMean(t *testing.T) {
	ev := []telemetry.Event{adv(), adv(), adv(), adv(), rej(), rej()}
	s := stepsToGreen(ev)
	if !s.HasValue || s.Mean != 1.5 {
		t.Fatalf("got %+v", s)
	}
}

// span returns min/max timestamps; empty/missing timestamps are ignored.
func TestSpanMinMax(t *testing.T) {
	ev := []telemetry.Event{
		{Timestamp: "2026-06-01T12:00:00Z"},
		{Timestamp: ""},
		{Timestamp: "2026-01-01T00:00:00Z"},
		{Timestamp: "2026-03-15T00:00:00Z"},
	}
	start, end := span(ev)
	if start != "2026-01-01T00:00:00Z" || end != "2026-06-01T12:00:00Z" {
		t.Fatalf("span = %q..%q", start, end)
	}
}

// No timestamps ⇒ empty span.
func TestSpanEmpty(t *testing.T) {
	start, end := span([]telemetry.Event{{Timestamp: ""}})
	if start != "" || end != "" {
		t.Fatalf("span = %q..%q, want empty", start, end)
	}
}
