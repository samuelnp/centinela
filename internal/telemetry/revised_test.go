package telemetry

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestRecordRevisedEvent(t *testing.T) {
	t.Chdir(t.TempDir())
	enabled := true
	cfg := &config.Config{}
	cfg.Telemetry = config.TelemetryConfig{Enabled: &enabled}

	RecordRevised(cfg, "f", "validate", "code", "opus")

	evs, err := ReadDefault()
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 {
		t.Fatalf("events = %d, want 1", len(evs))
	}
	e := evs[0]
	if e.Type != TypeStepRevised {
		t.Fatalf("type = %q", e.Type)
	}
	// Step holds the target (to); From holds the step rewound away from.
	if e.From != "validate" || e.Step != "code" {
		t.Fatalf("endpoints = from %q to %q", e.From, e.Step)
	}
	if e.Feature != "f" || e.Model != "opus" {
		t.Fatalf("event = %+v", e)
	}
}
