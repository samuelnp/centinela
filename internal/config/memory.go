package config

const (
	// DefaultRecallMaxEntries caps how many ledger entries recall injects.
	DefaultRecallMaxEntries = 10
	// DefaultRecallMaxBytes caps the total injected byte budget.
	DefaultRecallMaxBytes = 4096
)

// MemoryConfig controls the governed project-memory ledger (capture + recall).
type MemoryConfig struct {
	// Enabled gates the whole subsystem. Default true.
	Enabled *bool `toml:"enabled"`
	// RecallMaxEntries caps recalled entries injected into the plan step.
	RecallMaxEntries int `toml:"recall_max_entries"`
	// RecallMaxBytes caps the total bytes of recalled entries.
	RecallMaxBytes int `toml:"recall_max_bytes"`
}

// IsEnabled reports whether memory capture/recall should run.
func (m MemoryConfig) IsEnabled() bool {
	return m.Enabled == nil || *m.Enabled
}

// NormalizeRecallMaxEntries clamps the entry cap to a positive default.
func NormalizeRecallMaxEntries(n int) int {
	if n <= 0 {
		return DefaultRecallMaxEntries
	}
	return n
}

// NormalizeRecallMaxBytes clamps the byte cap to a positive default.
func NormalizeRecallMaxBytes(n int) int {
	if n <= 0 {
		return DefaultRecallMaxBytes
	}
	return n
}

func applyMemoryDefaults(cfg *Config) {
	cfg.Memory.RecallMaxEntries = NormalizeRecallMaxEntries(cfg.Memory.RecallMaxEntries)
	cfg.Memory.RecallMaxBytes = NormalizeRecallMaxBytes(cfg.Memory.RecallMaxBytes)
}
