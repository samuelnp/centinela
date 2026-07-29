// Acceptance: specs/token-diet.feature
//
// The `evidence validate` CLI subcommand always passes nil uiPaths (it has no
// flag for them), so a ux-ui-specialist evidence file can never pass the
// "real UI file" output check through that surface. These tests call the
// exported orchestration.ValidateEvidence directly — the same function the
// CLI delegates to — supplying uiPaths the way `centinela validate` does
// internally, so the UX-tag rule itself is exercised end to end.
package acceptance_test

import (
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

const tdUIPath = "ui/widget.tsx"

// tdWriteUXEvidence writes a code-step ux-ui-specialist evidence file with the
// given edgeCases and mobileFirst, plus a real UI output file, and returns the
// validation error (nil on pass).
func tdWriteUXEvidence(t *testing.T, feature string, edgeCases []string, mobileFirst bool) error {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	mustWrite(t, tdUIPath, "export const Widget = () => null;\n")
	path := filepath.Join(".workflow", feature+"-ux-ui-specialist.json")
	body := `{"feature":"` + feature + `","step":"code","role":"ux-ui-specialist","status":"done",` +
		`"generatedAt":"2026-01-01T00:00:00Z","inputs":["x"],"outputs":["` + tdUIPath + `"],` +
		`"edgeCases":` + tdJSONArray(edgeCases) + `,"mobileFirst":` + boolStr(mobileFirst) + `,"handoffTo":"qa-senior"}`
	mustWrite(t, path, body)
	return orchestration.ValidateEvidence(path, feature, "code", orchestration.RoleUXUISpecialist, []string{"ui/"})
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// Scenario: A descriptive UX edge case satisfies its required tag
func TestTD_DescriptiveUXEdgeCaseSatisfiesRequiredTag(t *testing.T) {
	edgeCases := []string{
		"mobile-first: renders at 80x24",
		"visual-hierarchy: primary action stands out",
		"typography-hierarchy: headings scale down",
		"responsive-layout: reflows under 320px",
		"loading-state: spinner shown while pending",
		"empty-state: friendly copy with no rows",
		"error-state: retry button on failure",
		"motion-and-reduced-motion: respects prefers-reduced-motion",
	}
	if err := tdWriteUXEvidence(t, "token-diet", edgeCases, true); err != nil {
		t.Fatalf("descriptive tags must satisfy all required tags: %v", err)
	}
}

// Scenario: Bare tags keep working exactly as before
func TestTD_BareUXTagsKeepWorking(t *testing.T) {
	bare := []string{
		"mobile-first", "visual-hierarchy", "typography-hierarchy", "responsive-layout",
		"loading-state", "empty-state", "error-state", "motion-and-reduced-motion",
	}
	if err := tdWriteUXEvidence(t, "token-diet", bare, true); err != nil {
		t.Fatalf("bare tags must still validate (strict loosening only): %v", err)
	}
}
