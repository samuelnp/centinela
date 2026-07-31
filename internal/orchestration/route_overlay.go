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

// CanonicalRouteTiers folds raw route keys onto normalized role slugs. The state
// file is agent-writable, so two raw keys ("Senior-Engineer" and
// "senior-engineer") can name the same role; resolving them in map order would
// let Go's randomized iteration pick the winner, and the same state would emit a
// different model run to run. Unknown roles and ambiguous collisions are
// DROPPED — deterministic by construction, and the same fail-safe posture a
// corrupt tier already gets.
func CanonicalRouteTiers(routes map[string]string) map[string]string {
	if len(routes) == 0 {
		return nil
	}
	out := make(map[string]string, len(routes))
	collided := map[string]bool{}
	for raw, tier := range routes {
		role, known := NormalizeRole(raw)
		if !known {
			continue
		}
		if _, seen := out[string(role)]; seen {
			collided[string(role)] = true
			continue
		}
		out[string(role)] = tier
	}
	for slug := range collided {
		delete(out, slug)
	}
	return out
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
	canonical := CanonicalRouteTiers(routes)
	if len(canonical) == 0 {
		return models
	}
	out := make(RoleModels, len(models)+len(canonical))
	for slug, rm := range models {
		out[slug] = rm
	}
	for slug, tier := range canonical {
		role := Role(slug)
		routed, honored := HonoredRouteTier(role, tier, floors)
		if !honored || routed == RoleTier(role, models) {
			continue
		}
		out[slug] = RoleModel{Tier: string(routed)}
	}
	return out
}
