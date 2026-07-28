package orchestration

import "strings"

// Tier is a semantic model class that shields user config from concrete,
// per-runner model IDs (which drift on every model release).
type Tier string

const (
	TierReasoning Tier = "reasoning"
	TierBalanced  Tier = "balanced"
	TierFast      Tier = "fast"
)

// defaultTierForRole encodes the locked role→tier defaults. The step roles are
// emitted by the orchestration hook; the remaining out-of-band roles
// (edge-case-tester, merge-steward) are documented here for completeness but
// are not injected by the directive in v1.
var defaultTierForRole = map[Role]Tier{
	RoleBigThinker:     TierReasoning,
	RoleSeniorEngineer: TierReasoning,
	// Refutation is the hardest judgment in the workflow, and the verifier now
	// replaces the validation-specialist context outright, so the net subagent
	// spend still drops (D7).
	RoleGatekeeper:       TierReasoning,
	RoleFeatureSpecial:   TierBalanced,
	RoleQASeniorEngineer: TierBalanced,
	RoleUXUISpecialist:   TierBalanced,
	RoleDocsSpecialist:   TierFast,
	RoleValidationSpec:   TierFast,
	RoleMergeSteward:     TierReasoning,
	// Out-of-band role not declared in policy.go as a constant.
	Role("edge-case-tester"): TierFast,
}

// DefaultTierForRole returns the built-in tier for a role, or TierBalanced for
// any role without an explicit default.
func DefaultTierForRole(role Role) Tier {
	if tier, ok := defaultTierForRole[role]; ok {
		return tier
	}
	return TierBalanced
}

// NormalizeTier trims and lowercases s, then validates it against the allowed
// tiers. ok is false for any unrecognized value.
func NormalizeTier(s string) (Tier, bool) {
	candidate := Tier(strings.ToLower(strings.TrimSpace(s)))
	for _, tier := range AllowedTiers() {
		if candidate == tier {
			return tier, true
		}
	}
	return "", false
}

// AllowedTiers returns the valid tiers in a stable order.
func AllowedTiers() []Tier {
	return []Tier{TierReasoning, TierBalanced, TierFast}
}

// AllowedRunnerKeys returns the runner keys config may key model overrides on,
// in stable order. RunnerUnknown is internal and never a valid config key.
func AllowedRunnerKeys() []string {
	return []string{string(RunnerClaude), string(RunnerOpenCode), string(RunnerCodex)}
}

// AllowedRoleSlugs returns the role slugs config may key on: the 7 step roles
// plus the out-of-band roles that carry a documented default tier.
func AllowedRoleSlugs() []string {
	return []string{
		string(RoleBigThinker),
		string(RoleFeatureSpecial),
		string(RoleSeniorEngineer),
		string(RoleUXUISpecialist),
		string(RoleQASeniorEngineer),
		string(RoleDocsSpecialist),
		string(RoleValidationSpec),
		string(RoleMergeSteward),
		string(RoleGatekeeper),
		"edge-case-tester",
	}
}
