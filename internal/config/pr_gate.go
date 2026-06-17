package config

// PrGateConfig controls the `centinela pr-gate` changed-since-base gate run.
// Enabled governs only the CI advisory surface; the command runs whenever
// invoked. FailOnWarning (default false) escalates a warn-severity gate to a
// non-zero exit so a PR can be blocked on warnings when an operator opts in.
type PrGateConfig struct {
	Enabled       bool `toml:"enabled"`
	FailOnWarning bool `toml:"fail_on_warning"`
}

// NormalizePrGate is a no-op today: both fields default to their safe zero
// values (disabled advisory surface, fail_on_warning = false).
func NormalizePrGate(c PrGateConfig) PrGateConfig {
	return c
}

// validatePrGate is reserved for future knobs; currently a no-op.
func validatePrGate(_ PrGateConfig) error {
	return nil
}
