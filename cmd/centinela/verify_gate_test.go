package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// writeGateRepo chdirs into a temp repo with the given validate command and a
// qa-senior evidence file, returning a loaded config.
func writeGateRepo(t *testing.T, validateCmd, evidence string) *config.Config {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile("centinela.toml", []byte("[validate]\ncommands=[\""+validateCmd+"\"]\n"), 0o644)
	_ = os.MkdirAll(workflow.WorkflowDir, 0o755)
	wf := workflow.New("feat")
	wf.CurrentStep = "validate"
	_ = workflow.Save(wf)
	_ = os.WriteFile(".workflow/feat-validation-specialist.json", []byte(evidence), 0o644)
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

const valEvidence = `{"feature":"feat","step":"validate","role":"validation-specialist","status":"done","generatedAt":"2026-05-29T00:00:00Z","inputs":["i"],"outputs":[],"edgeCases":[],"handoffTo":"orchestrator"}`

func TestRunClaimVerificationHardBlocksOnFailedTests(t *testing.T) {
	cfg := writeGateRepo(t, "false", valEvidence)
	err := runClaimVerification("feat", "validate", cfg)
	if err == nil {
		t.Fatal("failing validate.commands must hard-block completion")
	}
	if !strings.Contains(err.Error(), "diverges from ground truth") {
		t.Fatalf("error should name claim divergence, got %v", err)
	}
}

func TestRunClaimVerificationPassesOnHonestEvidence(t *testing.T) {
	cfg := writeGateRepo(t, "true", valEvidence)
	if err := runClaimVerification("feat", "validate", cfg); err != nil {
		t.Fatalf("honest passing evidence should not block: %v", err)
	}
}
