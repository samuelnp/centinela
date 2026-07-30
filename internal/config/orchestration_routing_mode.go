package config

import (
	"fmt"
	"strings"
)

// RoutingModeStatic is the DEFAULT and the value an absent key resolves to:
// every command and hook emits exactly today's bytes. RoutingModeDynamic opts
// the project into per-feature model routing via `centinela route`.
const (
	RoutingModeStatic  = "static"
	RoutingModeDynamic = "dynamic"
)

// DynamicRoutingEnabled reports whether [orchestration] routing_mode selects
// dynamic routing. A nil config, an absent key, and "static" all return false,
// so the static path is never entered through a new branch by accident.
func DynamicRoutingEnabled(cfg *Config) bool {
	if cfg == nil {
		return false
	}
	return normalizeRoutingMode(cfg.Orchestration.RoutingMode) == RoutingModeDynamic
}

// normalizeRoutingMode trims + lowercases the raw value; empty means absent,
// which resolves to the static default.
func normalizeRoutingMode(raw string) string {
	mode := strings.ToLower(strings.TrimSpace(raw))
	if mode == "" {
		return RoutingModeStatic
	}
	return mode
}

// validateRoutingMode rejects anything that is neither static nor dynamic. A
// typo must fail at load rather than silently degrade to static.
func validateRoutingMode(cfg *Config) error {
	switch normalizeRoutingMode(cfg.Orchestration.RoutingMode) {
	case RoutingModeStatic, RoutingModeDynamic:
		return nil
	}
	return fmt.Errorf("orchestration.routing_mode: unknown mode %q (allowed: %s, %s)",
		cfg.Orchestration.RoutingMode, RoutingModeStatic, RoutingModeDynamic)
}

// OrchestrationFloors returns role→minimum-tier from [orchestration.floors],
// normalized (trim + lowercase) into a fresh map. Retired plan-role keys alias
// onto "planner" exactly as they do for [orchestration.models] (D8). Semantics
// — shipped defaults, ordering, refusal — live in internal/orchestration; this
// leaf only publishes the configured values.
func OrchestrationFloors(cfg *Config) map[string]string {
	if cfg == nil || len(cfg.Orchestration.Floors) == 0 {
		return nil
	}
	out := make(map[string]string, len(cfg.Orchestration.Floors))
	for roleKey, tier := range cfg.Orchestration.Floors {
		out[strings.ToLower(strings.TrimSpace(roleKey))] = strings.ToLower(strings.TrimSpace(tier))
	}
	aliasPlanFloor(out)
	return out
}

// aliasPlanFloor republishes a legacy plan-role floor under "planner" when no
// explicit planner floor is configured. The sibling of aliasPlanRole for a map
// whose values are plain strings rather than RoleModelValue unions.
func aliasPlanFloor(out map[string]string) {
	if _, ok := out[plannerRoleKey]; ok {
		return
	}
	for _, key := range legacyPlanRoleKeys {
		if tier, ok := out[key]; ok {
			out[plannerRoleKey] = tier
			return
		}
	}
}

// validateOrchestrationFloors rejects unknown role keys and invalid tier values
// in [orchestration.floors], reusing the SAME local sets [orchestration.models]
// validates against so the two vocabularies can never drift.
func validateOrchestrationFloors(cfg *Config) error {
	for roleKey, tierValue := range cfg.Orchestration.Floors {
		role := strings.ToLower(strings.TrimSpace(roleKey))
		if !allowedModelRoles[role] {
			return fmt.Errorf("orchestration.floors: unknown role key %q", roleKey)
		}
		tier := strings.ToLower(strings.TrimSpace(tierValue))
		if !allowedModelTiers[tier] {
			return fmt.Errorf("orchestration.floors[%q]: invalid tier %q (allowed: %s)",
				roleKey, tierValue, allowedTiersList())
		}
	}
	return nil
}
