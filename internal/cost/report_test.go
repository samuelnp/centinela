package cost

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/telemetry"
)

func sampleEvents() []telemetry.Event {
	return []telemetry.Event{
		{Type: telemetry.TypeCostSample, Feature: "f", Step: "code", Model: "m1", InputTokens: 600, OutputTokens: 400},
		{Type: telemetry.TypeCostSample, Feature: "f", Step: "code", Model: "m1", InputTokens: 100, OutputTokens: 0},
		{Type: telemetry.TypeStepAdvanced, Feature: "f", Step: "code"}, // ignored
	}
}

func TestFoldOnlyCostSamples(t *testing.T) {
	a := Fold(sampleEvents())
	if got := a.Feature["f"].Tokens(); got != 1100 {
		t.Fatalf("feature total = %d, want 1100", got)
	}
	if got := a.Step["f"]["code"].Tokens(); got != 1100 {
		t.Fatalf("step total = %d, want 1100", got)
	}
	if got := a.Model["m1"].Tokens(); got != 1100 {
		t.Fatalf("model total = %d, want 1100", got)
	}
}

func TestBudgetStatusOverUnderAndRemaining(t *testing.T) {
	over := status("step", "f/code", 1100, 1000)
	if !over.Over || over.Remaining() != 0 {
		t.Fatalf("over budget: %+v", over)
	}
	under := status("feature", "f", 1100, 5000)
	if under.Over || under.Remaining() != 3900 {
		t.Fatalf("under budget: %+v rem=%d", under, under.Remaining())
	}
	none := status("model", "m1", 1100, 0) // 0 budget never warns
	if none.Over || none.Remaining() != 0 {
		t.Fatalf("zero budget should never be over: %+v", none)
	}
}

func TestBuildReportAndQueries(t *testing.T) {
	cfg := config.CostConfig{Enabled: true, StepTokenBudget: 1000, FeatureTokenBudget: 5000}
	r := Build(Fold(sampleEvents()), cfg)
	if r.Empty() {
		t.Fatal("report should not be empty")
	}
	if !r.AnyOver() { // step 1100 > 1000
		t.Fatal("expected an over-budget row")
	}
	st, over := ActiveStatus(Fold(sampleEvents()), cfg, "f", "code")
	if !over || st.Scope != "step" {
		t.Fatalf("active step should be over budget, got %+v over=%v", st, over)
	}
}

func TestEmptyReportAndNoActiveOver(t *testing.T) {
	r := Build(Fold(nil), config.CostConfig{})
	if !r.Empty() || r.AnyOver() {
		t.Fatalf("nil aggregate → empty, not over: %+v", r)
	}
	if _, over := ActiveStatus(Fold(nil), config.CostConfig{}, "x", "y"); over {
		t.Fatal("no samples → never over budget")
	}
}
