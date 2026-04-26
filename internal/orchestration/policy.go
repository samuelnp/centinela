package orchestration

type Role string

const (
	RoleBigThinker       Role = "big-thinker"
	RoleFeatureSpecial   Role = "feature-specialist"
	RoleSeniorEngineer   Role = "senior-engineer"
	RoleUXUISpecialist   Role = "ux-ui-specialist"
	RoleQASeniorEngineer Role = "qa-senior"
	RoleDocsSpecialist   Role = "documentation-specialist"
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
	default:
		return nil
	}
}

func RequiredRolesForFeature(feature, step string) []Role {
	roles := append([]Role{}, RequiredRoles(step)...)
	if step == "code" && IsUserFacingFeature(feature) {
		roles = append(roles, RoleUXUISpecialist)
	}
	return roles
}
