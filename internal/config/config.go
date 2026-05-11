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
	Gates         GatesConfig         `toml:"gates"`
	I18n          I18nConfig          `toml:"i18n"`
}

// ValidateConfig holds user-defined commands that centinela runs during validate.
type ValidateConfig struct {
	Commands []string `toml:"commands"`
	DiffMode string   `toml:"diff_mode"`
	DiffBase string   `toml:"diff_base"`
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
	cfg.Workflow.PlanAdvisorMode = NormalizePlanAdvisorMode(cfg.Workflow.PlanAdvisorMode)
	cfg.Workflow.PlanQuestionLimit = NormalizePlanQuestionLimit(cfg.Workflow.PlanQuestionLimit)
	cfg.Validate.DiffMode = NormalizeDiffMode(cfg.Validate.DiffMode)
	cfg.Validate.DiffBase = NormalizeDiffBase(cfg.Validate.DiffBase)
}
