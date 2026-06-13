// Acceptance: specs/right-size-docs-step.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// rdsBrief writes a feature brief with the given body under a temp CWD.
func rdsBrief(t *testing.T, feature, body string) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("docs/features", feature+".md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Scenario: A user-facing feature still requires the documentation-specialist role
func TestRDSUserFacingRequiresDocsSpecialist(t *testing.T) {
	rdsBrief(t, "uf", "# uf\nsurface: user-facing\n")
	roles := orchestration.RequiredRolesForFeature("uf", "docs")
	if !slices.Contains(roles, orchestration.RoleDocsSpecialist) {
		t.Fatalf("user-facing docs step must require documentation-specialist, got %v", roles)
	}
}

// Scenario: An internal feature does not require the documentation-specialist role
func TestRDSInternalDropsDocsSpecialist(t *testing.T) {
	rdsBrief(t, "in", "# in\nsurface: internal\n")
	roles := orchestration.RequiredRolesForFeature("in", "docs")
	if slices.Contains(roles, orchestration.RoleDocsSpecialist) {
		t.Fatalf("internal docs step must not require documentation-specialist, got %v", roles)
	}
}

// Scenario: The default surface is internal when none is declared
func TestRDSDefaultSurfaceIsInternal(t *testing.T) {
	rdsBrief(t, "none", "# none\nno surface line here\n")
	if orchestration.IsUserFacingFeature("none") {
		t.Fatal("a brief with no surface line must be treated as internal")
	}
}
