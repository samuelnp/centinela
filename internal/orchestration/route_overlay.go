package orchestration

// HonoredRouteTier resolves the tier a RECORDED route actually takes effect at.
//
// Floors are enforced HERE, on the resolution path, not only in ValidateRoute:
// the workflow state file is agent-writable in every step, so a hand-written
// route below the role's floor would otherwise silently downgrade the role —
// including the gatekeeper, the one role whose tier IS the quality of the
// adversarial judgment. A route that fails the floor is treated exactly like an
// unparseable tier: IGNORED, never fatal, falling back to the static resolution.
func HonoredRouteTier(role Role, tier string, floors map[string]string) (Tier, bool) {
	normalized, ok := NormalizeTier(tier)
	if !ok {
		return "", false
	}
	if floor, hasFloor := EffectiveFloor(role, floors); hasFloor && TierBelow(normalized, floor) {
		return "", false
	}
	return normalized, true
}

// ApplyRoutes returns models with each HONORED routed tier replacing that role's
// entry — routes-first, so a per-feature decision beats a project-wide role
// override table. The input map is never mutated: the hook shares one models map
// across every active workflow, and each workflow's routes are its own.
//
// A route to the role's CURRENT effective tier is a no-op and leaves the static
// entry (including its per-runner concrete-model override table) untouched: an
// operator confirming the tier the directive already printed must never silently
// discard a pin. A route to a DIFFERENT tier legitimately replaces the entry.
func ApplyRoutes(models RoleModels, routes, floors map[string]string) RoleModels {
	if len(routes) == 0 {
		return models
	}
	out := make(RoleModels, len(models)+len(routes))
	for slug, rm := range models {
		out[slug] = rm
	}
	for slug, tier := range routes {
		role, known := NormalizeRole(slug)
		if !known {
			continue
		}
		routed, honored := HonoredRouteTier(role, tier, floors)
		if !honored || routed == RoleTier(role, models) {
			continue
		}
		out[string(role)] = RoleModel{Tier: string(routed)}
	}
	return out
}
