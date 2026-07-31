package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const uxEdgeCases = `"mobile-first","visual-hierarchy","typography-hierarchy",` +
	`"responsive-layout","loading-state","empty-state","error-state",` +
	`"motion-and-reduced-motion"`

// writeUXEvidence writes a contract-valid ux-ui-specialist artifact whose only
// variable is the single output path, so the UX-output rule (the one that
// consults uiPaths) is the only thing that can fail.
func writeUXEvidence(t *testing.T, feature, output string) {
	t.Helper()
	if dir := filepath.Dir(output); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(output, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body := `{"feature":"` + feature + `","step":"code","role":"ux-ui-specialist",` +
		`"status":"done","generatedAt":"2026-01-01T00:00:00Z",` +
		`"inputs":["docs/features/` + feature + `.md"],"outputs":["` + output + `"],` +
		`"edgeCases":[` + uxEdgeCases + `],"mobileFirst":true,"handoffTo":"qa-senior"}`
	path := filepath.Join(".workflow", feature+"-ux-ui-specialist.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Red-before-green: under the previous nil uiPaths the UX-output rule could
// never match, so a ux role with a genuinely UI output was permanently red
// from the CLI while `complete` passed it.
func TestEvidenceValidate_UXOutputUnderConfiguredUIPath_IsClean(t *testing.T) {
	chdirEvidenceTemp(t)
	if err := os.WriteFile("centinela.toml",
		[]byte("[orchestration]\nui_paths = [\"web/app\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeUXEvidence(t, "alpha", "web/app/panel.tsx")
	if err := runEvidenceValidate(nil, []string{"alpha"}); err != nil {
		t.Fatalf("configured ui path must satisfy the UX-output rule, got %v", err)
	}
}

// The mirror direction: an output touching no UI path still reports the issue.
func TestEvidenceValidate_UXOutputOutsideUIPaths_ReportsHint(t *testing.T) {
	chdirEvidenceTemp(t)
	if err := os.WriteFile("centinela.toml",
		[]byte("[orchestration]\nui_paths = [\"web/app\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeUXEvidence(t, "alpha", "docs/notes.md")
	var validateErr error
	stderr := captureStderr(t, func() {
		validateErr = runEvidenceValidate(nil, []string{"alpha"})
	})
	if validateErr == nil {
		t.Fatal("expected a UX-output issue for a non-UI output")
	}
	if !strings.Contains(stderr, "ux-ui-specialist outputs must include") {
		t.Fatalf("expected the UX-output hint on stderr, got %q", stderr)
	}
}

// E1: no centinela.toml at all — config.Load returns defaults and the rule is
// evaluated against the built-in default UI paths (internal/ui is one).
func TestEvidenceValidate_NoConfigFile_UsesDefaultUIPaths(t *testing.T) {
	chdirEvidenceTemp(t)
	writeUXEvidence(t, "alpha", "internal/ui/panel.go")
	if err := runEvidenceValidate(nil, []string{"alpha"}); err != nil {
		t.Fatalf("default ui paths must apply with no config file, got %v", err)
	}
}
