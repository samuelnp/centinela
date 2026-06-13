package acceptance_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/telemetry"
)

// Acceptance: specs/governance-telemetry.feature

func gtWriteRaw(t *testing.T, lines string) {
	t.Helper()
	if err := os.MkdirAll(".workflow/telemetry", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(".workflow/telemetry", "events.jsonl"), []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Scenario: Multiple events accumulate append-only in call order
func TestGT_AccumulateInOrder(t *testing.T) {
	gtChdir(t)
	cfg := gtCfg(true)
	telemetry.RecordBlock(cfg, "f", "plan", "tests", "/p/tests/a_test.go", "out-of-step")
	telemetry.RecordGateFailure(cfg, "G", "m")
	telemetry.RecordStepAdvanced(cfg, "f", "validate")
	evs := gtEvents(t)
	if len(evs) != 3 || evs[0].Type != telemetry.TypeBlock ||
		evs[1].Type != telemetry.TypeGateFailure || evs[2].Type != telemetry.TypeStepAdvanced {
		t.Fatalf("events not in append order: %+v", evs)
	}
}

// Scenario: Two sequential records both land intact under append-only writes
func TestGT_TwoSequentialIntact(t *testing.T) {
	gtChdir(t)
	cfg := gtCfg(true)
	telemetry.RecordStepAdvanced(cfg, "f", "plan")
	telemetry.RecordStepAdvanced(cfg, "f", "code")
	evs := gtEvents(t)
	if len(evs) != 2 || evs[0].Step != "plan" || evs[1].Step != "code" {
		t.Fatalf("both sequential records must land intact: %+v", evs)
	}
}

// Scenario: An I/O error while recording does not fail the host command
func TestGT_IOErrorDoesNotFail(t *testing.T) {
	gtChdir(t)
	// A regular file where the .workflow dir must be makes MkdirAll fail.
	if err := os.WriteFile(".workflow", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	telemetry.RecordGateFailure(gtCfg(true), "G", "m") // must return without panic
	if _, err := os.Stat(".workflow/telemetry/events.jsonl"); err == nil {
		t.Fatal("no events file should exist after an I/O error")
	}
}

// Scenario: Read skips a corrupt line and returns the valid events
func TestGT_ReadSkipsCorruptLine(t *testing.T) {
	gtChdir(t)
	gtWriteRaw(t, `{"schema":"centinela.telemetry/v1","type":"block"}
not valid json
{"schema":"centinela.telemetry/v1","type":"step-advanced"}
`)
	evs := gtEvents(t)
	if len(evs) != 2 || evs[0].Type != telemetry.TypeBlock || evs[1].Type != telemetry.TypeStepAdvanced {
		t.Fatalf("Read must skip the corrupt line and keep valid events: %+v", evs)
	}
}

// Scenario: Read of a missing telemetry log returns no events and no error
func TestGT_ReadMissingLog(t *testing.T) {
	gtChdir(t)
	evs, err := telemetry.ReadDefault()
	if err != nil || len(evs) != 0 {
		t.Fatalf("missing log must yield (nil,nil), got %d events err=%v", len(evs), err)
	}
}
