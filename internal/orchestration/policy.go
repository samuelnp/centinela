package orchestration

type Role string

const (
	RoleBigThinker       Role = "big-thinker"
	RoleFeatureSpecial   Role = "feature-specialist"
	RoleSeniorEngineer   Role = "senior-engineer"
	RoleQASeniorEngineer Role = "qa-senior"
)

func RequiredRoles(step string) []Role {
	switch step {
	case "plan":
		return []Role{RoleBigThinker, RoleFeatureSpecial}
	case "code":
		return []Role{RoleSeniorEngineer}
	case "tests":
		return []Role{RoleQASeniorEngineer}
	default:
		return nil
	}
}
