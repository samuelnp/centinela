// Acceptance: specs/truthful-validators.feature
//
// Section A — `centinela evidence validate` applies the UX-output rule using
// the configured (or default) UI paths instead of a permanently-nil list.
// Every scenario is driven through the compiled binary in a scratch dir; no
// fixture uses a network remote.
package acceptance_test

import (
	"strings"
	"testing"
)

const tvUXEdgeCases = `"mobile-first","visual-hierarchy","typography-hierarchy",` +
	`"responsive-layout","loading-state","empty-state","error-state",` +
	`"motion-and-reduced-motion"`

// tvWriteUXEvidence writes a contract-valid ux-ui-specialist evidence file
// whose only variable is the output path, so the UX-output rule is the only
// thing that can fail.
func tvWriteUXEvidence(t *testing.T, dir, feature, output string) {
	t.Helper()
	writeFile(t, dir, output, "x\n")
	body := `{"feature":"` + feature + `","step":"code","role":"ux-ui-specialist",` +
		`"status":"done","generatedAt":"2026-01-01T00:00:00Z",` +
		`"inputs":["docs/features/` + feature + `.md"],"outputs":["` + output + `"],` +
		`"edgeCases":[` + tvUXEdgeCases + `],"mobileFirst":true,"handoffTo":"qa-senior"}`
	writeFile(t, dir, ".workflow/"+feature+"-ux-ui-specialist.json", body)
}

// Scenario: Evidence that fails the UX-output rule under complete also fails from the CLI
func TestTV_EvidenceValidate_UXOutputOutsideUIPaths_Fails(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	tvWriteUXEvidence(t, dir, "alpha", "docs/notes.md")
	out, code := runCent(t, bin, dir, "evidence", "validate", "alpha")
	if code == 0 {
		t.Fatalf("expected non-zero exit for a UX output outside every UI path\n%s", out)
	}
	mustContain(t, out, "ux-ui-specialist outputs must include")
}

// Scenario: Evidence whose outputs include a configured UI path validates
func TestTV_EvidenceValidate_UXOutputUnderConfiguredUIPath_Passes(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	writeFile(t, dir, "centinela.toml", "[orchestration]\nui_paths = [\"web/app\"]\n")
	tvWriteUXEvidence(t, dir, "alpha", "web/app/panel.tsx")
	out, code := runCent(t, bin, dir, "evidence", "validate", "alpha")
	if code != 0 {
		t.Fatalf("configured ui path must satisfy the rule\n%s", out)
	}
	mustContain(t, out, "evidence ok")
}

// Scenario: A repository with no centinela.toml still validates evidence
func TestTV_EvidenceValidate_NoConfigFile_UsesBuiltinDefaults(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	tvWriteUXEvidence(t, dir, "alpha", "internal/ui/panel.go")
	out, code := runCent(t, bin, dir, "evidence", "validate", "alpha")
	if code != 0 {
		t.Fatalf("no centinela.toml must still validate against built-in defaults\n%s", out)
	}
}

// Scenario: An unparseable centinela.toml is a reported error, not a silent downgrade
func TestTV_EvidenceValidate_UnparseableConfig_IsReportedError(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	writeFile(t, dir, "centinela.toml", "[orchestration\nui_paths = ")
	tvWriteUXEvidence(t, dir, "alpha", "web/panel.tsx")
	out, code := runCent(t, bin, dir, "evidence", "validate", "alpha")
	if code == 0 {
		t.Fatalf("a malformed centinela.toml must exit non-zero\n%s", out)
	}
	mustContain(t, out, "centinela.toml")
	if strings.Contains(out, "evidence ok") {
		t.Fatalf("a config error must never report evidence ok\n%s", out)
	}
}

// Scenario: An explicitly empty ui_paths list falls back to the built-in defaults
func TestTV_EvidenceValidate_EmptyUIPaths_FallsBackToDefaults(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	writeFile(t, dir, "centinela.toml", "[orchestration]\nui_paths = []\n")
	tvWriteUXEvidence(t, dir, "alpha", "web/panel.tsx")
	out, code := runCent(t, bin, dir, "evidence", "validate", "alpha")
	if code != 0 {
		t.Fatalf("an explicitly empty ui_paths list must fall back to defaults\n%s", out)
	}
}
