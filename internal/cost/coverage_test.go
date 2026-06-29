package cost

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/telemetry"
)

// ActiveStatus falls through to the feature budget when the step is within
// budget (or absent) but the feature total is over.
func TestActiveStatusFeatureBranch(t *testing.T) {
	ev := []telemetry.Event{
		{Type: telemetry.TypeCostSample, Feature: "f", Step: "code", InputTokens: 6000, OutputTokens: 0},
	}
	cfg := config.CostConfig{Enabled: true, StepTokenBudget: 100000, FeatureTokenBudget: 5000}
	st, over := ActiveStatus(Fold(ev), cfg, "f", "code")
	if !over || st.Scope != "feature" {
		t.Fatalf("expected feature-scope over budget, got %+v over=%v", st, over)
	}
}

// A step with no recorded samples skips the step branch entirely.
func TestActiveStatusStepAbsent(t *testing.T) {
	ev := []telemetry.Event{
		{Type: telemetry.TypeCostSample, Feature: "f", Step: "plan", InputTokens: 10, OutputTokens: 0},
	}
	cfg := config.CostConfig{Enabled: true, StepTokenBudget: 1, FeatureTokenBudget: 0}
	if _, over := ActiveStatus(Fold(ev), cfg, "f", "code"); over {
		t.Fatal("absent step with no feature budget must not be over")
	}
}

func TestSortByNameOrders(t *testing.T) {
	s := []Status{{Name: "z"}, {Name: "a"}, {Name: "m"}}
	sortByName(s)
	if s[0].Name != "a" || s[1].Name != "m" || s[2].Name != "z" {
		t.Fatalf("not sorted: %+v", s)
	}
}
