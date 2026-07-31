package evidence

import (
	"encoding/json"
	"time"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// stepForRole maps each role to the workflow step it belongs to so the
// skeleton always starts schema-valid for that role.
func stepForRole(role Role) string {
	switch role {
	case orchestration.RolePlanner, orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial:
		return "plan"
	case orchestration.RoleSeniorEngineer, orchestration.RoleUXUISpecialist:
		return "code"
	case orchestration.RoleQASeniorEngineer:
		return "tests"
	case orchestration.RoleValidationSpec, orchestration.RoleGatekeeper, Role("production-readiness"):
		return "validate"
	case orchestration.RoleDocsSpecialist:
		return "docs"
	case orchestration.RoleMergeSteward:
		return "merge"
	default:
		return ""
	}
}

// handoffForRole returns the successor the FEATURE'S OWN workflow contract
// derives, so a freshly scaffolded stub is never seeded with a value the chain
// gate would then refuse — the CLI's own prefill failing the CLI's own gate is
// the foot-gun this delegation removes.
//
// merge-steward is out-of-band (its step never appears in the ordered steps),
// so its literal verdict pair is answered here rather than derived.
//
// The static defaults below remain the fallback for a feature with no workflow
// state on disk: there is no contract to derive from, and a stub authored
// outside a workflow is not gated by one either.
func handoffForRole(feature string, role Role) string {
	if role == orchestration.RoleMergeSteward {
		return "complete"
	}
	if want, ok := workflow.ExpectedHandoff(feature, stepForRole(role), role); ok && want != "" {
		return want
	}
	return legacyHandoffForRole(role)
}

// legacyHandoffForRole is the pre-derivation static chain, kept verbatim as
// the no-workflow-state fallback.
func legacyHandoffForRole(role Role) string {
	switch role {
	case orchestration.RolePlanner:
		return string(orchestration.RoleSeniorEngineer)
	case orchestration.RoleBigThinker:
		return string(orchestration.RoleFeatureSpecial)
	case orchestration.RoleFeatureSpecial:
		return string(orchestration.RoleSeniorEngineer)
	case orchestration.RoleSeniorEngineer, orchestration.RoleUXUISpecialist:
		return string(orchestration.RoleQASeniorEngineer)
	case orchestration.RoleQASeniorEngineer:
		return string(orchestration.RoleValidationSpec)
	case orchestration.RoleGatekeeper, orchestration.RoleValidationSpec:
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
		HandoffTo:   handoffForRole(feature, role),
		Extra:       map[string]json.RawMessage{},
	}
	if role == orchestration.RoleUXUISpecialist {
		t := true
		r.MobileFirst = &t
	}
	return r
}
