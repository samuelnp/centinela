package evidence

import (
	"encoding/json"
	"time"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// stepForRole maps each role to the workflow step it belongs to so the
// skeleton always starts schema-valid for that role.
func stepForRole(role Role) string {
	switch role {
	case orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial:
		return "plan"
	case orchestration.RoleSeniorEngineer, orchestration.RoleUXUISpecialist:
		return "code"
	case orchestration.RoleQASeniorEngineer:
		return "tests"
	case orchestration.RoleValidationSpec, Role("gatekeeper"), Role("production-readiness"):
		return "validate"
	case orchestration.RoleDocsSpecialist:
		return "docs"
	case orchestration.RoleMergeSteward:
		return "merge"
	default:
		return ""
	}
}

// handoffForRole returns the canonical next role per the contract.
func handoffForRole(role Role) string {
	switch role {
	case orchestration.RoleBigThinker:
		return string(orchestration.RoleFeatureSpecial)
	case orchestration.RoleFeatureSpecial:
		return string(orchestration.RoleSeniorEngineer)
	case orchestration.RoleSeniorEngineer, orchestration.RoleUXUISpecialist:
		return string(orchestration.RoleQASeniorEngineer)
	case orchestration.RoleQASeniorEngineer:
		return string(orchestration.RoleValidationSpec)
	case orchestration.RoleValidationSpec:
		return string(orchestration.RoleDocsSpecialist)
	case orchestration.RoleDocsSpecialist, orchestration.RoleMergeSteward:
		return "complete"
	default:
		return "complete"
	}
}

// Skeleton returns a fresh RoleEvidence pre-populated with a Meta block and
// the role/step pair the contract requires. Required list fields are left
// empty so `centinela evidence append` is the authoring path.
func Skeleton(feature string, role Role, cliVersion string) *RoleEvidence {
	r := &RoleEvidence{
		Meta:        &Meta{CLIVersion: cliVersion, WrittenAt: time.Now().UTC().Format(time.RFC3339)},
		Feature:     feature,
		Step:        stepForRole(role),
		Role:        string(role),
		Status:      "done",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Inputs:      []string{},
		Outputs:     []string{},
		EdgeCases:   []string{},
		HandoffTo:   handoffForRole(role),
		Extra:       map[string]json.RawMessage{},
	}
	if role == orchestration.RoleUXUISpecialist {
		t := true
		r.MobileFirst = &t
	}
	return r
}
