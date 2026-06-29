package config

// CostConfig controls the optional cost-governance soft gate. When active,
// Centinela attributes host-harness token spend to the active feature/step and
// surfaces a non-blocking warning once spend exceeds a budget. All budgets are
// token counts (0 = no budget for that scope). For local models the same counts
// stand in for a compute/wall-clock unit. This gate NEVER blocks.
type CostConfig struct {
	Enabled            bool           `toml:"enabled"`
	StepTokenBudget    int            `toml:"step_token_budget"`
	FeatureTokenBudget int            `toml:"feature_token_budget"`
	TierBudgets        map[string]int `toml:"tier_budgets"`
}

// NormalizeCost clamps negative budgets to 0 (off) so a typo cannot invert the
// comparison. Defaults are all-zero: an unconfigured [cost] is a silent no-op.
func NormalizeCost(c CostConfig) CostConfig {
	if c.StepTokenBudget < 0 {
		c.StepTokenBudget = 0
	}
	if c.FeatureTokenBudget < 0 {
		c.FeatureTokenBudget = 0
	}
	for k, v := range c.TierBudgets {
		if v < 0 {
			c.TierBudgets[k] = 0
		}
	}
	return c
}

// IsActive reports whether cost governance should do anything: it must be
// enabled AND have at least one positive budget. Zero config stays a no-op.
func (c CostConfig) IsActive() bool {
	if !c.Enabled {
		return false
	}
	if c.StepTokenBudget > 0 || c.FeatureTokenBudget > 0 {
		return true
	}
	for _, v := range c.TierBudgets {
		if v > 0 {
			return true
		}
	}
	return false
}
