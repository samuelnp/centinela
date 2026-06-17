package orchestration

import (
	"os"
	"slices"
	"testing"
)

// docsSurfaceFixture writes a feature brief with the given body and chdirs into
// a temp repo so IsUserFacingFeature reads it.
func docsSurfaceFixture(t *testing.T, feature, body string) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/features/"+feature+".md", []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRequiredRolesForFeatureDocsUserFacingRequiresDocsSpecialist(t *testing.T) {
	docsSurfaceFixture(t, "uf", "# uf\nsurface: user-facing\n")
	roles := RequiredRolesForFeature("uf", "docs")
	if !slices.Contains(roles, RoleDocsSpecialist) {
		t.Fatalf("user-facing docs step must require documentation-specialist, got %v", roles)
	}
}

func TestRequiredRolesForFeatureDocsInternalDropsDocsSpecialist(t *testing.T) {
	docsSurfaceFixture(t, "in", "# in\nsurface: internal\n")
	roles := RequiredRolesForFeature("in", "docs")
	if slices.Contains(roles, RoleDocsSpecialist) {
		t.Fatalf("internal docs step must not require documentation-specialist, got %v", roles)
	}
	if len(roles) != 0 {
		t.Fatalf("internal docs step should require no roles, got %v", roles)
	}
}

func TestRequiredRolesForFeatureCodeUXUnchangedBesideDocsGating(t *testing.T) {
	docsSurfaceFixture(t, "uf", "# uf\nsurface: user-facing\n")
	code := RequiredRolesForFeature("uf", "code")
	if !slices.Contains(code, RoleUXUISpecialist) {
		t.Fatalf("code step ux-ui gating must be unchanged, got %v", code)
	}
}
