package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// E3: an explicitly empty ui_paths list falls back to the built-in defaults
// (asserted through the CLI, not re-implemented — config.UIPaths owns it).
func TestEvidenceValidate_EmptyUIPathsFallsBackToDefaults(t *testing.T) {
	chdirEvidenceTemp(t)
	if err := os.WriteFile("centinela.toml",
		[]byte("[orchestration]\nui_paths = []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeUXEvidence(t, "alpha", "web/panel.tsx")
	if err := runEvidenceValidate(nil, []string{"alpha"}); err != nil {
		t.Fatalf("empty ui_paths must fall back to defaults, got %v", err)
	}
}

// E2: a malformed centinela.toml surfaces as a config error. It must never
// degrade to nil paths and must not print the success line.
func TestEvidenceValidate_UnparseableConfigIsAnError(t *testing.T) {
	chdirEvidenceTemp(t)
	if err := os.WriteFile("centinela.toml", []byte("[orchestration\nui_paths = "), 0o644); err != nil {
		t.Fatal(err)
	}
	writeUXEvidence(t, "alpha", "web/panel.tsx")
	out := captureStdout(t, func() {
		err := runEvidenceValidate(nil, []string{"alpha"})
		if err == nil || !strings.Contains(err.Error(), config.Filename) {
			t.Fatalf("expected a config parse error naming %s, got %v", config.Filename, err)
		}
	})
	if strings.Contains(out, "evidence ok") {
		t.Fatalf("a config error must not report evidence ok, got %q", out)
	}
}

// Parity: the CLI's verdict and the orchestration path's verdict for the same
// artifact agree, because both resolve uiPaths through config.UIPaths(cfg).
func TestEvidenceValidate_MatchesOrchestrationVerdict(t *testing.T) {
	chdirEvidenceTemp(t)
	if err := os.WriteFile("centinela.toml",
		[]byte("[orchestration]\nui_paths = [\"web/app\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		output   string
		wantFail bool
	}{
		{"web/app/panel.tsx", false},
		{"docs/notes.md", true},
	} {
		writeUXEvidence(t, "alpha", tc.output)
		cfg, err := config.Load()
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(".workflow", "alpha-ux-ui-specialist.json")
		orchErr := orchestration.ValidateEvidence(path, "alpha", "code",
			orchestration.RoleUXUISpecialist, config.UIPaths(cfg))
		var cliErr error
		_ = captureStderr(t, func() { cliErr = runEvidenceValidate(nil, []string{"alpha"}) })
		if (orchErr != nil) != tc.wantFail || (cliErr != nil) != tc.wantFail {
			t.Fatalf("output %q: orchestration=%v cli=%v want fail=%v",
				tc.output, orchErr, cliErr, tc.wantFail)
		}
	}
}
