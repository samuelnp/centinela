package config

// TelemetryConfig controls the governance event log (append-only JSONL).
type TelemetryConfig struct {
	// Enabled gates the whole subsystem. Default true (opt-out), like [memory].
	Enabled *bool `toml:"enabled"`
}

// IsEnabled reports whether telemetry recording should run.
func (t TelemetryConfig) IsEnabled() bool { return t.Enabled == nil || *t.Enabled }
