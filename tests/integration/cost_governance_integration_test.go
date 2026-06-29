package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/cost"
	"github.com/samuelnp/centinela/internal/telemetry"
)

// Integration: a recorded cost sample round-trips through the telemetry log and
// folds back into per-feature/step totals compared against budget.
func TestCostSampleRoundTripThroughTelemetry(t *testing.T) {
	t.Chdir(t.TempDir())
	telemetry.RecordCostSample(&config.Config{}, "cg", "code", "m1", 1200, 900)
	telemetry.RecordCostSample(&config.Config{}, "cg", "code", "m1", 100, 0)

	events, err := telemetry.ReadDefault()
	if err != nil {
		t.Fatal(err)
	}
	report := cost.Build(cost.Fold(events), config.CostConfig{
		Enabled: true, StepTokenBudget: 1000, FeatureTokenBudget: 5000,
	})
	if report.Empty() || !report.AnyOver() {
		t.Fatalf("expected a populated, over-budget report: %+v", report)
	}
	// The written log lives under the telemetry dir, not the repo root.
	if _, err := os.Stat(filepath.Join(".workflow", "telemetry", "events.jsonl")); err != nil {
		t.Fatalf("telemetry log should exist: %v", err)
	}
}
