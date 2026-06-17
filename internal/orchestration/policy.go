package orchestration

type Role string

const (
	RoleBigThinker       Role = "big-thinker"
	RoleFeatureSpecial   Role = "feature-specialist"
	RoleSeniorEngineer   Role = "senior-engineer"
	RoleUXUISpecialist   Role = "ux-ui-specialist"
	RoleQASeniorEngineer Role = "qa-senior"
	RoleDocsSpecialist   Role = "documentation-specialist"
	RoleValidationSpec   Role = "validation-specialist"
	// RoleMergeSteward runs out-of-band on `centinela merge <feature>` and is
	// not part of the 5-step workflow. It does not appear in RequiredRoles
	// for any step but it MUST validate as evidence when the merger writes
	// `.workflow/<feature>-merge-steward.{md,json}`.
	RoleMergeSteward Role = "merge-steward"
)

func RequiredRoles(step string) []Role {
	switch step {
	case "plan":
		return []Role{RoleBigThinker, RoleFeatureSpecial}
	case "code":
		return []Role{RoleSeniorEngineer}
	case "tests":
		return []Role{RoleQASeniorEngineer}
	case "docs":
		return []Role{RoleDocsSpecialist}
	case "validate":
		return []Role{RoleValidationSpec}
	default:
		return nil
	}
}

func RequiredRolesForFeature(feature, step string) []Role {
	if step == "docs" && !IsUserFacingFeature(feature) {
		// Internal features ship a one-line changelog instead of the full
		// knowledge-base bundle, so the documentation-specialist evidence is
		// not required. Mirrors the code-step ux-ui gating below.
		return nil
	}
	roles := append([]Role{}, RequiredRoles(step)...)
	if step == "code" && IsUserFacingFeature(feature) {
		roles = append(roles, RoleUXUISpecialist)
	}
	return roles
}
