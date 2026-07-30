package evidence

import "github.com/samuelnp/centinela/internal/orchestration"

// PlanInputs returns the mechanical inputs pre-fill for a role, or nil when the
// role derives nothing. Only plan roles snapshot docs — and only the feature's
// OWN brief and plan (two entries, no repo-wide glob). It shares
// orchestration.RequiredPlanInputs with the validator, so a pre-filled init
// satisfies validatePlanSnapshotInputs by construction.
func PlanInputs(feature string, role Role) []string {
	switch role {
	case orchestration.RolePlanner, orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial:
		return orchestration.RequiredPlanInputs(feature)
	default:
		return nil
	}
}
