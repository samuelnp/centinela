package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/calibration"
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/telemetry"
)

// calCfg returns a telemetry-enabled config that also maps the three spec model
// ids to their capability classes (so bare ids classify without the built-in
// id-spelling dependency).
func calCfg() *config.Config {
	c := &config.Config{}
	c.Orchestration.Capabilities = map[string]string{
		"claude-opus-4-7":   config.CapabilityFrontier,
		"claude-sonnet-4-6": config.CapabilityCapable,
		"claude-haiku-4-5":  config.CapabilityLimited,
	}
	return c
}

// repeatRec invokes rec n times.
func repeatRec(n int, rec func()) {
	for i := 0; i < n; i++ {
		rec()
	}
}

// Full pipeline: stamp events with models via the Record* constructors, Read them
// back from the temp dir, Calibrate, and assert each model's classification.
func TestCalibrationStampReadCalibrateRoundTrip(t *testing.T) {
	t.Chdir(t.TempDir())
	cfg := calCfg()

	// Sonnet: 3 advances + 3 rework → rate 1.0 → Undergoverned/Tighten.
	repeatRec(3, func() { telemetry.RecordStepAdvanced(cfg, "f", "plan", "claude-sonnet-4-6") })
	repeatRec(3, func() { telemetry.RecordGateFailure(cfg, "G1", "too big", "claude-sonnet-4-6") })
	// Haiku: 4 advances + 1 rework → rate 0.25 → Overgoverned/Loosen.
	repeatRec(4, func() { telemetry.RecordStepAdvanced(cfg, "f", "code", "claude-haiku-4-5") })
	telemetry.RecordGateFailure(cfg, "G1", "x", "claude-haiku-4-5")
	// Unattributed (empty model) event.
	telemetry.RecordStepAdvanced(cfg, "f", "plan", "")

	events, err := telemetry.Read(filepath.Join(".workflow", "telemetry"))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(events) != 12 {
		t.Fatalf("expected 12 events, got %d", len(events))
	}

	rep := calibration.Calibrate(events, cfg)
	if rep.ModelCount != 3 {
		t.Fatalf("ModelCount = %d, want 3", rep.ModelCount)
	}
	got := map[string]calibration.Verdict{}
	for _, m := range rep.Models {
		got[m.Model] = m.Verdict
	}
	if got["claude-sonnet-4-6"] != calibration.Undergoverned {
		t.Fatalf("sonnet verdict = %v, want Undergoverned", got["claude-sonnet-4-6"])
	}
	if got["claude-haiku-4-5"] != calibration.Overgoverned {
		t.Fatalf("haiku verdict = %v, want Overgoverned", got["claude-haiku-4-5"])
	}
	if got["unattributed"] != calibration.Unclassified {
		t.Fatalf("unattributed verdict = %v, want Unclassified", got["unattributed"])
	}
	if rep.Models[len(rep.Models)-1].Model != "unattributed" {
		t.Fatalf("unattributed should sort last: %+v", rep.Models)
	}
}
