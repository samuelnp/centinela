package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/calibration"
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/telemetry"
)

// Acceptance: specs/capability-calibration.feature

// calEnabledCfg returns a config with telemetry enabled (default-on).
func calEnabledCfg() *config.Config { return &config.Config{} }

// Scenario: Event recorded during a workflow with a pinned DriverModel carries that model in the JSONL
func TestCalStampingCarriesModel(t *testing.T) {
	t.Chdir(t.TempDir())
	// The resolved driver model (resolveEmitModel result) is passed to the
	// constructor by the cmd/ caller; here we pass the pinned model directly.
	telemetry.RecordStepAdvanced(calEnabledCfg(), "alpha", "plan", "claude-sonnet-4-6")
	evs, err := telemetry.ReadDefault()
	if err != nil || len(evs) != 1 {
		t.Fatalf("read: %v len=%d", err, len(evs))
	}
	if evs[0].Model != "claude-sonnet-4-6" {
		t.Fatalf("model not stamped on JSONL: %+v", evs[0])
	}
}

// Scenario: Event recorded with no driver model configured has an empty model field
func TestCalStampingEmptyModel(t *testing.T) {
	t.Chdir(t.TempDir())
	telemetry.RecordStepAdvanced(calEnabledCfg(), "alpha", "plan", "")
	evs, err := telemetry.ReadDefault()
	if err != nil || len(evs) != 1 {
		t.Fatalf("read: %v len=%d", err, len(evs))
	}
	if evs[0].Model != "" {
		t.Fatalf("expected empty model, got %q", evs[0].Model)
	}
}

// Scenario: Legacy event without a model field parses cleanly and buckets as unattributed
func TestCalStampingLegacyBucketsUnattributed(t *testing.T) {
	dir := calRepo(t, []string{
		`{"schema":"centinela.telemetry/v1","type":"step-advanced","timestamp":"2026-01-01T00:00:00Z","feature":"a","step":"plan"}`,
	})
	evs, err := telemetry.Read(dir + "/.workflow/telemetry")
	if err != nil || len(evs) != 1 {
		t.Fatalf("read: %v len=%d", err, len(evs))
	}
	if evs[0].Model != "" {
		t.Fatalf("legacy Model should be empty, got %q", evs[0].Model)
	}
	rep := calibration.Calibrate(evs, nil)
	if rep.ModelCount != 1 || rep.Models[0].Model != "unattributed" {
		t.Fatalf("legacy event should bucket unattributed: %+v", rep.Models)
	}
}
