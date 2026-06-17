package config

import (
	"fmt"
	"strings"
)

// allowedModelTiers and allowedModelRoles are LOCAL string sets: the config
// leaf may NOT import internal/orchestration. A cross-package parity test keeps
// these in sync with the domain's AllowedTiers()/AllowedRoleSlugs().
var allowedModelTiers = map[string]bool{
	"reasoning": true,
	"balanced":  true,
	"fast":      true,
}

var allowedModelRoles = map[string]bool{
	"big-thinker":              true,
	"feature-specialist":       true,
	"senior-engineer":          true,
	"ux-ui-specialist":         true,
	"qa-senior":                true,
	"documentation-specialist": true,
	"validation-specialist":    true,
	"merge-steward":            true,
	"gatekeeper":               true,
	"edge-case-tester":         true,
}

// validateOrchestrationModels rejects unknown role keys, invalid tier values
// (string form), and malformed runner→model tables (override form) in
// [orchestration.models]. An absent/empty table is valid.
func validateOrchestrationModels(cfg *Config) error {
	for roleKey, value := range cfg.Orchestration.Models {
		if !allowedModelRoles[roleKey] {
			return fmt.Errorf("orchestration.models: unknown role key %q", roleKey)
		}
		if err := validateRoleModelValue(roleKey, value); err != nil {
			return err
		}
	}
	return nil
}

// validateRoleModelValue validates one union entry: the tier string form, or
// the runner→model override table form.
func validateRoleModelValue(roleKey string, value RoleModelValue) error {
	if len(value.Overrides) > 0 {
		for runnerKey, model := range value.Overrides {
			runner := strings.ToLower(strings.TrimSpace(runnerKey))
			if !allowedRunnerKeys[runner] {
				return fmt.Errorf("orchestration.models[%q]: unknown runner key %q (allowed: %s)",
					roleKey, runnerKey, allowedRunnerKeysList())
			}
			if strings.TrimSpace(model) == "" {
				return fmt.Errorf("orchestration.models[%q].%s: model string must not be empty", roleKey, runnerKey)
			}
		}
		return nil
	}
	tier := strings.ToLower(strings.TrimSpace(value.Tier))
	if !allowedModelTiers[tier] {
		return fmt.Errorf("orchestration.models[%q]: invalid tier %q (allowed: %s)",
			roleKey, value.Tier, allowedTiersList())
	}
	return nil
}

func allowedTiersList() string {
	return strings.Join([]string{"reasoning", "balanced", "fast"}, ", ")
}
