package config

// plannerRoleKey is the [orchestration.models] key the unified plan role uses.
const plannerRoleKey = "planner"

// legacyPlanRoleKeys are the retired plan-role config keys in PRECEDENCE
// order: big-thinker set the reasoning tier that planner inherits, so it wins
// over feature-specialist when both are present and planner is absent.
var legacyPlanRoleKeys = []string{"big-thinker", "feature-specialist"}

// aliasPlanRole republishes a legacy plan-role entry from out under "planner"
// when the config carries no explicit planner entry in EITHER value form (D8).
// Checking models rather than out is what keeps an explicit `planner = "fast"`
// from being shadowed by an aliased `big-thinker = { ... }` override table.
// The legacy keys are LEFT in out so existing consumers and their tests are
// untouched — this only adds, never rewrites or removes.
func aliasPlanRole[T any](models map[string]RoleModelValue, out map[string]T) {
	if out == nil {
		return
	}
	if _, ok := models[plannerRoleKey]; ok {
		return
	}
	for _, key := range legacyPlanRoleKeys {
		if v, ok := out[key]; ok {
			out[plannerRoleKey] = v
			return
		}
	}
}

// LegacyPlanModelRoleKeys returns the retired plan-role keys present in
// [orchestration.models], sorted in the documented precedence order. It is a
// pure, nil-safe read used to surface a ONE-TIME migration notice at
// `centinela doctor` and `centinela start` — never on the per-prompt hook path,
// where it would become permanent noise.
func LegacyPlanModelRoleKeys(cfg *Config) []string {
	if cfg == nil || len(cfg.Orchestration.Models) == 0 {
		return nil
	}
	var out []string
	for _, key := range legacyPlanRoleKeys {
		if _, ok := cfg.Orchestration.Models[key]; ok {
			out = append(out, key)
		}
	}
	return out
}

// LegacyPlanModelRoleNotice renders the deprecation line for the keys, or "" when
// there are none. Centralized so doctor and start print the identical wording.
func LegacyPlanModelRoleNotice(cfg *Config) string {
	keys := LegacyPlanModelRoleKeys(cfg)
	if len(keys) == 0 {
		return ""
	}
	list := keys[0]
	for _, k := range keys[1:] {
		list += ", " + k
	}
	return "[orchestration.models] uses retired plan role key(s): " + list +
		` — rename to "planner" (the legacy key still resolves, aliased to planner)`
}
