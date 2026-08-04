package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// TestExpectedHandoffDerivationMatrix replaces the plan's prose argument with
// a table: every documented hop, plus the two cases a hardcoded five-role
// chain got WRONG (an internal feature's validate step is terminal because its
// docs step requires no role, and a legacy pin keeps its own occupant).
func TestExpectedHandoffDerivationMatrix(t *testing.T) {
	runHandoffMatrix(t, []hoCase{
		{name: "internal plan", step: "plan",
			role: orchestration.RolePlanner, want: "senior-engineer"},
		{name: "internal code", step: "code",
			role: orchestration.RoleSeniorEngineer, want: "qa-senior"},
		{name: "internal tests", step: "tests",
			role: orchestration.RoleQASeniorEngineer, want: "gatekeeper"},
		{name: "internal validate is terminal", step: "validate",
			role: orchestration.RoleGatekeeper, want: TerminalHandoff},
		{name: "user-facing code hops within the step", userFacing: true, step: "code",
			role: orchestration.RoleSeniorEngineer, want: "ux-ui-specialist"},
		{name: "user-facing second code role", userFacing: true, step: "code",
			role: orchestration.RoleUXUISpecialist, want: "qa-senior"},
		{name: "user-facing validate reaches docs", userFacing: true, step: "validate",
			role: orchestration.RoleGatekeeper, want: "documentation-specialist"},
		{name: "user-facing docs is terminal", userFacing: true, step: "docs",
			role: orchestration.RoleDocsSpecialist, want: TerminalHandoff},
		{name: "legacy validate pin keeps its occupant", legacyValidate: true, step: "tests",
			role: orchestration.RoleQASeniorEngineer, want: "validation-specialist"},
		{name: "legacy validate pin is terminal too", legacyValidate: true, step: "validate",
			role: orchestration.RoleValidationSpec, want: TerminalHandoff},
		{name: "legacy plan pair hops within the step", legacyPlan: true, step: "plan",
			role: orchestration.RoleBigThinker, want: "feature-specialist"},
		{name: "legacy plan pair second role", legacyPlan: true, step: "plan",
			role: orchestration.RoleFeatureSpecial, want: "senior-engineer"},
	})
}

// Archetype subsets exercise the skip-empty-steps walk: each omits a different
// step, and the terminal "complete" must fall out of the order rather than be
// special-cased per archetype.
func TestExpectedHandoffAcrossArchetypeOrders(t *testing.T) {
	hotfix := []string{"code", "tests", "validate"}
	spike := []string{"plan", "code"}
	runHandoffMatrix(t, []hoCase{
		{name: "hotfix code", order: hotfix, step: "code",
			role: orchestration.RoleSeniorEngineer, want: "qa-senior"},
		{name: "hotfix validate is terminal", order: hotfix, step: "validate",
			role: orchestration.RoleGatekeeper, want: TerminalHandoff},
		{name: "spike plan", order: spike, step: "plan",
			role: orchestration.RolePlanner, want: "senior-engineer"},
		{name: "spike code is terminal", order: spike, step: "code",
			role: orchestration.RoleSeniorEngineer, want: TerminalHandoff},
		{name: "bootstrap code skips the absent tests step", order: BootstrapStepOrder, step: "code",
			role: orchestration.RoleSeniorEngineer, want: "gatekeeper"},
		{name: "bootstrap user-facing validate", userFacing: true, order: BootstrapStepOrder,
			step: "validate", role: orchestration.RoleGatekeeper, want: "documentation-specialist"},
	})
}

// The derivation is total for inputs outside the ordered steps — the
// out-of-band merge step has no successor, so it terminates rather than
// inventing one. Pinned so the totality is a stated behaviour, not a surprise.
func TestExpectedHandoffTerminatesOutsideTheOrderedSteps(t *testing.T) {
	runHandoffMatrix(t, []hoCase{
		{name: "merge step", step: "merge",
			role: orchestration.RoleMergeSteward, want: TerminalHandoff},
		{name: "unknown step", step: "nowhere",
			role: orchestration.RoleSeniorEngineer, want: TerminalHandoff},
	})
}

// No workflow state means no contract to derive from. Reporting ok == false
// rather than a default is what keeps the evidence-init prefill from seeding a
// value nothing on disk supports.
func TestExpectedHandoffReportsAbsentWorkflowState(t *testing.T) {
	t.Chdir(t.TempDir())
	if want, ok := ExpectedHandoff("ghost", "tests", orchestration.RoleQASeniorEngineer); ok {
		t.Fatalf("missing state must report absence, got %q", want)
	}
}
