package orchestration

// tierRank orders the tiers by capability so "below" is a total comparison. An
// unrecognized tier ranks 0, which keeps a corrupt value from ever reading as
// stronger than a real one.
var tierRank = map[Tier]int{
	TierReasoning: 3,
	TierBalanced:  2,
	TierFast:      1,
}

// TierBelow reports whether a is strictly weaker than b.
func TierBelow(a, b Tier) bool {
	return tierRank[a] < tierRank[b]
}

// defaultFloors are the shipped minimum tiers. The gatekeeper is the workflow's
// only adversarial judgment and the planner sets the whole feature's direction,
// so both are floored by default; every other role is floorless. A project may
// consciously lower either through [orchestration.floors].
var defaultFloors = map[Role]Tier{
	RoleGatekeeper: TierReasoning,
	RolePlanner:    TierBalanced,
}

// DefaultFloors returns a copy of the shipped role→minimum-tier defaults.
func DefaultFloors() map[Role]Tier {
	out := make(map[Role]Tier, len(defaultFloors))
	for role, tier := range defaultFloors {
		out[role] = tier
	}
	return out
}

// floorSuccessor maps a RETIRED role onto the role that replaced it, so a legacy
// workflow's plan/validate roles inherit their successor's floor instead of
// running floorless. internal/config already aliases retired config KEYS onto
// "planner" (D8); this is that alias in the direction a legacy workflow needs —
// without it, `big-thinker`/`feature-specialist`/`validation-specialist` are a
// standing bypass of the shipped planner and gatekeeper floors.
var floorSuccessor = map[Role]Role{
	RoleBigThinker:     RolePlanner,
	RoleFeatureSpecial: RolePlanner,
	RoleValidationSpec: RoleGatekeeper,
}

// EffectiveFloor resolves a role's minimum tier: an explicit, valid config entry
// REPLACES the shipped default (a project may lower it), else the default, else
// no floor at all. A retired role falls through to its successor's config entry
// and then to its successor's default. cfgFloors is the normalized map from
// config.OrchestrationFloors.
func EffectiveFloor(role Role, cfgFloors map[string]string) (Tier, bool) {
	candidates := []Role{role}
	if successor, ok := floorSuccessor[role]; ok {
		candidates = append(candidates, successor)
	}
	for _, candidate := range candidates {
		if tier, ok := NormalizeTier(cfgFloors[string(candidate)]); ok {
			return tier, true
		}
	}
	for _, candidate := range candidates {
		if tier, ok := defaultFloors[candidate]; ok {
			return tier, true
		}
	}
	return "", false
}
