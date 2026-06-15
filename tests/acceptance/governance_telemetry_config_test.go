package acceptance_test

import (
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/telemetry"
)

// Acceptance: specs/governance-telemetry.feature

// Scenario: Telemetry disabled is a no-op and writes no file
func TestGT_DisabledNoOp(t *testing.T) {
	gtChdir(t)
	telemetry.RecordStepAdvanced(gtCfg(false), "feat", "plan", "")
	if evs := gtEvents(t); len(evs) != 0 {
		t.Fatalf("disabled telemetry must write nothing, got %+v", evs)
	}
}

// Scenario: Absent telemetry config defaults to enabled and records events
func TestGT_DefaultEnabled(t *testing.T) {
	gtChdir(t)
	telemetry.RecordStepAdvanced(gtDefaultCfg(), "feat", "plan", "")
	if evs := gtEvents(t); len(evs) != 1 {
		t.Fatalf("absent config must default ON and record, got %d", len(evs))
	}
}

// Scenario: Every recorded event carries the schema id and an RFC3339 timestamp
func TestGT_SchemaAndTimestamp(t *testing.T) {
	gtChdir(t)
	telemetry.RecordGateFailure(gtCfg(true), "G", "m", "")
	e := gtEvents(t)[0]
	if e.Schema != telemetry.Schema {
		t.Fatalf("schema = %q, want %q", e.Schema, telemetry.Schema)
	}
	if _, err := time.Parse(time.RFC3339, e.Timestamp); err != nil {
		t.Fatalf("timestamp %q not RFC3339: %v", e.Timestamp, err)
	}
}
