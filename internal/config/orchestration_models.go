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

// validateOrchestrationModels rejects unknown role keys and invalid tier values
// in [orchestration.models]. Tiers are normalized (trim + lowercase) before
// validation. An absent/empty table is valid (defaults apply downstream).
func validateOrchestrationModels(cfg *Config) error {
	for roleKey, tierValue := range cfg.Orchestration.Models {
		if !allowedModelRoles[roleKey] {
			return fmt.Errorf("orchestration.models: unknown role key %q", roleKey)
		}
		tier := strings.ToLower(strings.TrimSpace(tierValue))
		if !allowedModelTiers[tier] {
			return fmt.Errorf("orchestration.models[%q]: invalid tier %q (allowed: %s)",
				roleKey, tierValue, allowedTiersList())
		}
	}
	return nil
}

func allowedTiersList() string {
	return strings.Join([]string{"reasoning", "balanced", "fast"}, ", ")
}
