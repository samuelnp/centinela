package config

// GatesConfig controls which built-in gates are active.
type GatesConfig struct {
	FileSizeEnabled            bool                   `toml:"file_size"`
	FileSizeExceptions         []FileSizeException    `toml:"file_size_exceptions"`
	I18nEnabled                bool                   `toml:"i18n"`
	ProductionReadinessEnabled bool                   `toml:"production_readiness"`
	Build                      BuildGateConfig        `toml:"build"`
	ImportGraph                ImportGraphConfig      `toml:"import_graph"`
	Security                   SecurityGateConfig     `toml:"security"`
	SpecTraceability           SpecTraceabilityConfig `toml:"spec_traceability"`
	RoadmapDrift               RoadmapDriftConfig     `toml:"roadmap_drift"`
	AuditBaseline              AuditBaselineConfig    `toml:"audit_baseline"`
	CustomGates                []CustomGate           `toml:"custom"`
}
