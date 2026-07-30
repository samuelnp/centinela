package orchestration

// ApplyRoutes returns models with each routed tier replacing that role's entry
// WHOLESALE — routes-first, so a per-feature decision beats a project-wide role
// override table. The input map is never mutated: the hook shares one models map
// across every active workflow, and each workflow's routes are its own.
//
// An unparseable tier (state corruption, or a route written by a newer binary)
// is IGNORED rather than fatal: routing is a cost knob and must never break the
// hook that emits the directive.
func ApplyRoutes(models RoleModels, routes map[string]string) RoleModels {
	if len(routes) == 0 {
		return models
	}
	out := make(RoleModels, len(models)+len(routes))
	for role, rm := range models {
		out[role] = rm
	}
	for role, tier := range routes {
		if normalized, ok := NormalizeTier(tier); ok {
			out[role] = RoleModel{Tier: string(normalized)}
		}
	}
	return out
}
