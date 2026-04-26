package orchestration

import (
	"os"
	"slices"
	"testing"
)

func TestRequiredRolesForFeatureAddsUXForUserFacingCode(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                              //nolint:errcheck
	os.Chdir(d)                                                                    //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                             //nolint:errcheck
	os.WriteFile("docs/features/ui.md", []byte("surface: user-facing\n"), 0644)    //nolint:errcheck
	os.WriteFile("docs/features/internal.md", []byte("surface: internal\n"), 0644) //nolint:errcheck
	uiRoles := RequiredRolesForFeature("ui", "code")
	if !slices.Contains(uiRoles, RoleSeniorEngineer) || !slices.Contains(uiRoles, RoleUXUISpecialist) {
		t.Fatalf("expected senior-engineer and ux-ui-specialist, got %v", uiRoles)
	}
	internalRoles := RequiredRolesForFeature("internal", "code")
	if slices.Contains(internalRoles, RoleUXUISpecialist) {
		t.Fatalf("did not expect ux-ui-specialist for internal feature: %v", internalRoles)
	}
}
