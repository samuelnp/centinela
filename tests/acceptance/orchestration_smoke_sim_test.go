package acceptance_test

import (
	"slices"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestOrchestrationSmokeSim_RequiresQASeniorForTestsStep(t *testing.T) {
	roles := orchestration.RequiredRoles("tests")
	if !slices.Contains(roles, orchestration.RoleQASeniorEngineer) {
		t.Fatalf("expected tests step to require %q, got %v", orchestration.RoleQASeniorEngineer, roles)
	}
	codeRoles := orchestration.RequiredRoles("code")
	if slices.Contains(codeRoles, orchestration.RoleQASeniorEngineer) {
		t.Fatalf("did not expect %q in code step roles: %v", orchestration.RoleQASeniorEngineer, codeRoles)
	}
}
