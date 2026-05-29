package config

// VerifyConfig tunes claim verification (centinela verify / the complete gate).
type VerifyConfig struct {
	// TimeoutSeconds bounds each re-run of a test command. Default 60.
	TimeoutSeconds int `toml:"verify_timeout"`
	// CoverageTolerance is the max fractional gap a coverage claim may exceed
	// measured coverage before the check fails (0.001 = 0.1%). Default 0.001.
	CoverageTolerance float64 `toml:"coverage_tolerance"`
}
