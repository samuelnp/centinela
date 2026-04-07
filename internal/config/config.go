package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

const Filename = "centinela.toml"

// Config is the centinela.toml structure.
type Config struct {
	Workflow WorkflowConfig `toml:"workflow"`
	Validate ValidateConfig `toml:"validate"`
	Gates    GatesConfig    `toml:"gates"`
	I18n     I18nConfig     `toml:"i18n"`
}

// WorkflowConfig controls language-specific step validation behaviour.
type WorkflowConfig struct {
	// TestSuffixes lists file suffixes that count as unit/integration tests.
	// If empty, any file in tests/unit or tests/integration is accepted.
	TestSuffixes []string `toml:"test_suffixes"`
	// AcceptanceSuffix is the file suffix for acceptance step definitions.
	// If empty, any file in tests/acceptance is accepted.
	AcceptanceSuffix string `toml:"acceptance_suffix"`
	// CodeDirs lists path segments that classify a file as "code".
	// If empty, a built-in set of common directories is used.
	CodeDirs []string `toml:"code_dirs"`
	// DisableAutoCommit prevents centinela from committing on step completion.
	DisableAutoCommit bool `toml:"disable_auto_commit"`
	// StepConfirmationMode controls when review prompts are shown.
	StepConfirmationMode string `toml:"step_confirmation_mode"`
}

// ValidateConfig holds user-defined commands that centinela runs during validate.
type ValidateConfig struct {
	Commands []string `toml:"commands"`
}

// GatesConfig controls which built-in gates are active.
type GatesConfig struct {
	FileSizeEnabled            bool                `toml:"file_size"`
	FileSizeExceptions         []FileSizeException `toml:"file_size_exceptions"`
	I18nEnabled                bool                `toml:"i18n"`
	ProductionReadinessEnabled bool                `toml:"production_readiness"`
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

	applyDefaults(&cfg)
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Gates: GatesConfig{
			FileSizeEnabled: true,
			I18nEnabled:     false,
		},
	}
}

func applyDefaults(cfg *Config) {
	if cfg.Gates.FileSizeEnabled == false && cfg.Gates.I18nEnabled == false {
		cfg.Gates.FileSizeEnabled = true
	}
	cfg.Workflow.StepConfirmationMode = NormalizeStepConfirmationMode(cfg.Workflow.StepConfirmationMode)
}
