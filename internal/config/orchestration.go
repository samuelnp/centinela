package config

import (
	"path/filepath"
	"strings"
)

type OrchestrationConfig struct {
	UIPaths            []string                     `toml:"ui_paths"`
	Models             map[string]RoleModelValue    `toml:"models"`
	ModelMap           map[string]map[string]string `toml:"model_map"`
	Capabilities       map[string]string            `toml:"capabilities"`        // model id → capability class
	CapabilityProfiles map[string]string            `toml:"capability_profiles"` // class → enforcement profile
	DriverModel        string                       `toml:"driver_model"`        // workflow's default-profile key
	Local              LocalConfig                  `toml:"local"`               // local-model backend (lowest-precedence tier)
}

// RoleModelValue is the union value of an [orchestration.models].<role> entry:
// EITHER a plain tier string (back-compat) OR a runner→model table (override).
// Exactly one of Tier / Overrides is populated after unmarshal.
type RoleModelValue struct {
	Tier      string            // set when the role value is a plain tier string
	Overrides map[string]string // set when the role value is a runner→model table
}

// OrchestrationModels is the back-compat accessor: it returns role→tier for
// the plain-string form (configurable-subagent-models). Role-override (table)
// entries are omitted. Equivalent to OrchestrationModelTiers.
func OrchestrationModels(cfg *Config) map[string]string {
	return OrchestrationModelTiers(cfg)
}

// OrchestrationModelTiers returns role→tier for entries given as a plain tier
// string (the back-compat form). Role-override (table) entries are omitted.
func OrchestrationModelTiers(cfg *Config) map[string]string {
	if cfg == nil {
		return nil
	}
	out := map[string]string{}
	for role, val := range cfg.Orchestration.Models {
		if val.Tier != "" {
			out[role] = val.Tier
		}
	}
	return out
}

// OrchestrationModelOverrides returns role→(runner→model) for entries given as
// a runner→model table (the role-override form).
func OrchestrationModelOverrides(cfg *Config) map[string]map[string]string {
	if cfg == nil {
		return nil
	}
	out := map[string]map[string]string{}
	for role, val := range cfg.Orchestration.Models {
		if len(val.Overrides) > 0 {
			out[role] = val.Overrides
		}
	}
	return out
}

// OrchestrationModelMap returns the tier→runner→model remap table (nil-safe).
func OrchestrationModelMap(cfg *Config) map[string]map[string]string {
	if cfg == nil {
		return nil
	}
	return cfg.Orchestration.ModelMap
}

var defaultUIPaths = []string{"src/ui", "src/components", "app/views", "web", "styles", "internal/ui"}

func UIPaths(cfg *Config) []string {
	paths := defaultUIPaths
	if cfg != nil && len(cfg.Orchestration.UIPaths) > 0 {
		paths = cfg.Orchestration.UIPaths
	}
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		clean := strings.TrimSpace(strings.TrimPrefix(path, "./"))
		if clean != "" {
			out = append(out, filepath.ToSlash(filepath.Clean(clean)))
		}
	}
	return out
}
