package evidence

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// retiredPlanRoles are the two roles the unified planner replaced. They are
// still parseable (legacy evidence must remain readable and repairable) but a
// fresh stub may only be authored for a workflow that predates planner-v1.
var retiredPlanRoles = map[Role]bool{
	orchestration.RoleBigThinker:     true,
	orchestration.RoleFeatureSpecial: true,
}

// EnsureRoleAllowed reports whether a NEW evidence stub may be authored for
// role under feature (D7). The predicate is the same state-dated contract the
// plan gate uses, so the CLI never offers a stub the gate would then refuse —
// and never bricks an in-flight legacy workflow that must still write its pair.
//
// A missing or unreadable workflow is refused rather than allowed: without a
// state file there is no evidence the workflow predates the migration, and
// permitting the stub would reopen the forged-legacy-pair hole D2 closes.
func EnsureRoleAllowed(feature string, role Role) error {
	if !retiredPlanRoles[role] {
		return nil
	}
	if wf, err := workflow.Load(feature); err == nil && !wf.UsesUnifiedPlanner() {
		return nil
	}
	return fmt.Errorf("role %q is retired; use %q (the legacy role is only accepted for a workflow that predates %q)",
		role, orchestration.RolePlanner, workflow.PlanContractUnified)
}
