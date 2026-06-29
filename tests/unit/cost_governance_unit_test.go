package unit_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/cost"
	"github.com/samuelnp/centinela/internal/telemetry"
)

// Unit: the budget soft gate marks a scope over only when used exceeds a
// positive budget, and never for a zero (unset) budget.
func TestCostBudgetSoftGate(t *testing.T) {
	cfg := config.CostConfig{Enabled: true, StepTokenBudget: 1000, FeatureTokenBudget: 0}
	events := []telemetry.Event{
		{Type: telemetry.TypeCostSample, Feature: "f", Step: "code", InputTokens: 900, OutputTokens: 600},
	}
	agg := cost.Fold(events)
	st, over := cost.ActiveStatus(agg, cfg, "f", "code")
	if !over || st.Used != 1500 || st.Budget != 1000 {
		t.Fatalf("step should be over budget: %+v over=%v", st, over)
	}
	// Feature budget is 0 → never over even though spend is high.
	if _, fover := cost.ActiveStatus(agg, config.CostConfig{Enabled: true}, "f", "code"); fover {
		t.Fatal("zero budgets must never report over")
	}
}

// Unit: an unconfigured [cost] is inactive (silent no-op).
func TestCostInactiveByDefault(t *testing.T) {
	if (config.CostConfig{}).IsActive() {
		t.Fatal("default cost config must be inactive")
	}
}
