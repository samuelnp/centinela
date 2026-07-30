// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// Scenario: A tag that is genuinely absent is still reported missing
func TestTD_GenuinelyAbsentUXTagStillReportedMissing(t *testing.T) {
	edgeCases := []string{
		"mobile-first: ok", "visual-hierarchy: ok", "typography-hierarchy: ok",
		"responsive-layout: ok", "loading-state: ok", "empty-state: ok",
		"motion-and-reduced-motion: ok",
	}
	err := tdWriteUXEvidence(t, "token-diet", edgeCases, true)
	if err == nil {
		t.Fatal("omitting error-state must fail validation")
	}
	mustContain(t, err.Error(), "error-state")
	for _, covered := range []string{"mobile-first", "visual-hierarchy", "typography-hierarchy",
		"responsive-layout", "loading-state", "empty-state", "motion-and-reduced-motion"} {
		mustNotContain(t, err.Error(), covered)
	}
}

// Scenario: A non-UX role is unaffected by the matcher
func TestTD_NonUXRoleUnaffectedByMatcher(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	implFile := "src/widget.go"
	mustWrite(t, implFile, "package src\n")
	path := filepath.Join(".workflow", "token-diet-senior-engineer.json")
	body := `{"feature":"token-diet","step":"code","role":"senior-engineer","status":"done",` +
		`"generatedAt":"2026-01-01T00:00:00Z","inputs":["x"],"outputs":["` + implFile + `"],` +
		`"edgeCases":["whatever, no ux tags at all"],"handoffTo":"qa-senior"}`
	mustWrite(t, path, body)
	if err := orchestration.ValidateEvidence(path, "token-diet", "code", orchestration.RoleSeniorEngineer, nil); err != nil {
		t.Fatalf("senior-engineer evidence must not consult UX tag requirements: %v", err)
	}
	if _, err := os.Stat(implFile); err != nil {
		t.Fatal("sanity: implementation file must exist")
	}
}
