package unit_test

import (
	"slices"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestCodeRolesRemainBackendSafe(t *testing.T) {
	roles := orchestration.RequiredRoles("code")
	if !slices.Contains(roles, orchestration.RoleSeniorEngineer) {
		t.Fatalf("expected senior-engineer in code roles, got %v", roles)
	}
	if slices.Contains(roles, orchestration.RoleUXUISpecialist) {
		t.Fatalf("did not expect ux-ui-specialist in base code roles: %v", roles)
	}
}
