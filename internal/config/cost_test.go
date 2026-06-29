package config

import "testing"

func TestNormalizeCostClampsNegatives(t *testing.T) {
	c := NormalizeCost(CostConfig{
		StepTokenBudget:    -5,
		FeatureTokenBudget: -1,
		TierBudgets:        map[string]int{"m": -3, "n": 10},
	})
	if c.StepTokenBudget != 0 || c.FeatureTokenBudget != 0 {
		t.Fatalf("negatives should clamp to 0, got %+v", c)
	}
	if c.TierBudgets["m"] != 0 || c.TierBudgets["n"] != 10 {
		t.Fatalf("tier clamp wrong: %+v", c.TierBudgets)
	}
}

func TestCostIsActive(t *testing.T) {
	cases := []struct {
		name string
		c    CostConfig
		want bool
	}{
		{"disabled", CostConfig{Enabled: false, StepTokenBudget: 100}, false},
		{"enabled no budgets", CostConfig{Enabled: true}, false},
		{"enabled step budget", CostConfig{Enabled: true, StepTokenBudget: 1}, true},
		{"enabled feature budget", CostConfig{Enabled: true, FeatureTokenBudget: 1}, true},
		{"enabled tier budget", CostConfig{Enabled: true, TierBudgets: map[string]int{"m": 1}}, true},
		{"enabled zero tier", CostConfig{Enabled: true, TierBudgets: map[string]int{"m": 0}}, false},
	}
	for _, tc := range cases {
		if got := tc.c.IsActive(); got != tc.want {
			t.Errorf("%s: IsActive=%v want %v", tc.name, got, tc.want)
		}
	}
}
