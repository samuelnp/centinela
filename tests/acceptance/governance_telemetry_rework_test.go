package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/telemetry"
)

// Acceptance: specs/governance-telemetry.feature

// Scenario: Rework is derivable from two complete-rejected events before a step-advanced
func TestGT_ReworkDerivable(t *testing.T) {
	gtChdir(t)
	cfg := gtCfg(true)
	telemetry.RecordCompleteRejected(cfg, "f", "validate", "gates")
	telemetry.RecordCompleteRejected(cfg, "f", "validate", "verify")
	telemetry.RecordStepAdvanced(cfg, "f", "validate")
	rejects := 0
	for _, e := range gtEvents(t) {
		if e.Type == telemetry.TypeStepAdvanced && e.Step == "validate" {
			break
		}
		if e.Type == telemetry.TypeCompleteRejected && e.Feature == "f" && e.Step == "validate" {
			rejects++
		}
	}
	if rejects != 2 {
		t.Fatalf("rework count before advance = %d, want 2", rejects)
	}
}
