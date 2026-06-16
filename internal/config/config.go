package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

const Filename = "centinela.toml"

// Config is the centinela.toml structure.
type Config struct {
	Workflow      WorkflowConfig      `toml:"workflow"`
	Orchestration OrchestrationConfig `toml:"orchestration"`
	Validate      ValidateConfig      `toml:"validate"`
	Verify        VerifyConfig        `toml:"verify"`
	Gates         GatesConfig         `toml:"gates"`
	I18n          I18nConfig          `toml:"i18n"`
	Memory        MemoryConfig        `toml:"memory"`
	Telemetry     TelemetryConfig     `toml:"telemetry"`
	Headless      HeadlessConfig      `toml:"headless"`
}

// ValidateConfig holds user-defined commands that centinela runs during validate.
type ValidateConfig struct {
	Commands []string `toml:"commands"`
	DiffMode string   `toml:"diff_mode"`
	DiffBase string   `toml:"diff_base"`
}

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

// I18nConfig describes how to check translations for G11.
type I18nConfig struct {
	// Format: "json" | "gettext" | "none"
	Format string `toml:"format"`
	// Dir is the directory containing locale files.
	Dir string `toml:"dir"`
	// Locales lists expected locale codes (e.g. ["en", "es"]).
	Locales []string `toml:"locales"`
}

// Load reads centinela.toml from the current directory.
func Load() (*Config, error) {
	data, err := os.ReadFile(Filename)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig(), nil
		}
		return nil, fmt.Errorf("reading %s: %w", Filename, err)
	}

	var cfg Config
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", Filename, err)
	}

	// Capture the explicit-vs-defaulted signal before applyDefaults overwrites
	// the empty step_confirmation_mode with every_step (see RawStepConfirmationMode).
	cfg.Workflow.RawStepConfirmationMode = cfg.Workflow.StepConfirmationMode
	// Same explicit-vs-defaulted capture for the enforcement profile: record the
	// raw value before applyDefaults normalizes the empty knob to strict, so the
	// capability tier can engage only when no global profile was explicitly set.
	cfg.Workflow.RawEnforcementProfile = cfg.Workflow.EnforcementProfile
	// Reject an explicitly-set unknown profile against the raw decoded value,
	// before applyDefaults normalizes it to strict (which would hide the error).
	if err := validateEnforcementProfile(cfg.Workflow.EnforcementProfile); err != nil {
		return nil, err
	}
	applyDefaults(&cfg)
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func defaultConfig() *Config {
	cfg := &Config{
		Gates: GatesConfig{
			FileSizeEnabled: true,
			I18nEnabled:     false,
		},
	}
	applyMemoryDefaults(cfg)
	return cfg
}
