package config

import (
	"fmt"
	"strings"
)

// allowedRunnerKeys is a LOCAL string set mirroring the domain's
// AllowedRunnerKeys(); the config leaf may NOT import internal/orchestration. A
// cross-package parity test keeps it in sync.
var allowedRunnerKeys = map[string]bool{
	"claude":   true,
	"opencode": true,
	"codex":    true,
}

// UnmarshalTOML decodes the union [orchestration.models].<role> value: a plain
// tier string (back-compat) OR a runner→model table (role override).
func (v *RoleModelValue) UnmarshalTOML(data any) error {
	switch raw := data.(type) {
	case string:
		v.Tier = raw
		return nil
	case map[string]any:
		overrides := make(map[string]string, len(raw))
		for runner, model := range raw {
			s, ok := model.(string)
			if !ok {
				return fmt.Errorf("orchestration.models: model for runner %q must be a string", runner)
			}
			overrides[runner] = s
		}
		v.Overrides = overrides
		return nil
	default:
		return fmt.Errorf("orchestration.models: value must be a tier string or a runner→model table")
	}
}

// validateOrchestrationModelMap rejects unknown tiers, unknown runner keys, and
// empty model strings in [orchestration.model_map]. Tier and runner keys are
// normalized (trim + lowercase) before validation. An absent/empty table is
// valid (defaults apply downstream).
func validateOrchestrationModelMap(cfg *Config) error {
	for tierKey, byRunner := range cfg.Orchestration.ModelMap {
		tier := strings.ToLower(strings.TrimSpace(tierKey))
		if !allowedModelTiers[tier] {
			return fmt.Errorf("orchestration.model_map: unknown tier key %q (allowed: %s)", tierKey, allowedTiersList())
		}
		for runnerKey, model := range byRunner {
			runner := strings.ToLower(strings.TrimSpace(runnerKey))
			if !allowedRunnerKeys[runner] {
				return fmt.Errorf("orchestration.model_map[%q]: unknown runner key %q (allowed: %s)",
					tierKey, runnerKey, allowedRunnerKeysList())
			}
			if strings.TrimSpace(model) == "" {
				return fmt.Errorf("orchestration.model_map[%q].%s: model string must not be empty", tierKey, runnerKey)
			}
		}
	}
	return nil
}

func allowedRunnerKeysList() string {
	return strings.Join([]string{"claude", "opencode", "codex"}, ", ")
}
