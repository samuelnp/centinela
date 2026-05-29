package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// setupVerifyDir chdirs into a temp repo with a started feature at the tests
// step and the given qa-senior evidence JSON written.
func setupVerifyDir(t *testing.T, evidence string) {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile("centinela.toml", []byte("[validate]\ncommands=[\"true\"]\n"), 0o644)
	_ = os.MkdirAll(workflow.WorkflowDir, 0o755)
	wf := workflow.New("feat")
	wf.CurrentStep = "tests"
	_ = workflow.Save(wf)
	if evidence != "" {
		_ = os.WriteFile(".workflow/feat-qa-senior.json", []byte(evidence), 0o644)
	}
}

const honestEvidence = `{"feature":"feat","step":"tests","role":"qa-senior","status":"done","generatedAt":"2026-05-29T00:00:00Z","inputs":["i"],"outputs":[],"edgeCases":[],"handoffTo":"validation-specialist"}`

func TestRunVerifyNoEvidenceClean(t *testing.T) {
	setupVerifyDir(t, "")
	if err := runVerify(nil, []string{"feat"}); err != nil {
		t.Fatalf("no evidence should verify clean: %v", err)
	}
}

func TestRunVerifyHonestPasses(t *testing.T) {
	setupVerifyDir(t, honestEvidence)
	if err := runVerify(nil, []string{"feat"}); err != nil {
		t.Fatalf("honest evidence should pass: %v", err)
	}
}

func TestRunVerifyMissingWorkflow(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	_ = os.Chdir(d)
	if err := runVerify(nil, []string{"nope"}); err == nil {
		t.Fatal("expected error for missing workflow")
	}
}

func TestVerifyRoot(t *testing.T) {
	if verifyRoot() == "" {
		t.Fatal("verifyRoot should never be empty")
	}
}
