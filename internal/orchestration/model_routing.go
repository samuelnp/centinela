package orchestration

// RoleModel is the resolved per-role routing config: an optional tier override
// (back-compat plain-string form) and an optional runner→concrete-model table
// (the role-override form). Both may be empty when the role is unconfigured.
type RoleModel struct {
	Tier      string            // tier name (may be empty → use the role's default tier)
	Overrides map[string]string // runner → concrete model ID (role-level override)
}

// RoleModels maps a role slug to its resolved routing config.
type RoleModels map[string]RoleModel

// ModelMap is the tier-remap table: tier → runner → concrete model ID.
type ModelMap map[string]map[string]string

// ResolveModel resolves a role to its concrete model ID for the given runner,
// applying the 4-step precedence:
//  1. role override [orchestration.models].<role>.<runner>
//  2. role→tier then tier-map override [orchestration.model_map].<tier>.<runner>
//  3. built-in tier→model default for the runner
//  4. missing mapping → return the TIER NAME with ok=false (never another
//     runner's concrete ID, never a crash)
func ResolveModel(role Role, models RoleModels, modelMap ModelMap, runner Runner) (string, bool) {
	rm := models[string(role)]
	// Step 1 — role-level override for this runner wins outright.
	if id, ok := rm.Overrides[string(runner)]; ok && id != "" {
		return id, true
	}
	tier := roleTier(role, rm.Tier)
	// Step 2 — tier-map override for this runner.
	if id, ok := modelMap[string(tier)][string(runner)]; ok && id != "" {
		return id, true
	}
	// Step 3 — built-in default for this runner.
	if id, ok := tierModels[tier][runner]; ok && id != "" {
		return id, true
	}
	// Step 4 — no mapping for the active runner: tier name + ok=false.
	return string(tier), false
}

// RoleTier returns the effective tier for a role given its routing config:
// the explicit tier override when present and valid, else the built-in default.
func RoleTier(role Role, models RoleModels) Tier {
	return roleTier(role, models[string(role)].Tier)
}

// roleTier resolves the effective tier for a role: an explicit, valid tier
// override when present, otherwise the role's built-in default tier.
func roleTier(role Role, override string) Tier {
	if override != "" {
		if normalized, ok := NormalizeTier(override); ok {
			return normalized
		}
	}
	return DefaultTierForRole(role)
}
