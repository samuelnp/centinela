package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// directiveFor runs the orchestration hook for a single strict workflow at the
// validate step pinned to the given contract, and returns its output.
func directiveFor(t *testing.T, contract string) string {
	t.Helper()
	t.Chdir(t.TempDir())
	os.MkdirAll(workflow.WorkflowDir, 0o755) //nolint:errcheck
	wf := workflow.New("f")
	wf.ValidateContract = contract
	wf.CurrentStep = "validate"
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
	return captureStdout(t, func() {
		withStdin(t, "{}", func() { runHookOrchestration(nil, nil) }) //nolint:errcheck
	})
}

// Regression: the directive and the completion gate resolved required roles
// through DIFFERENT functions, so a legacy strict workflow was told to write
// gatekeeper evidence while `complete` demanded validation-specialist.
func TestHookDirectiveNamesValidationSpecialistForLegacyWorkflow(t *testing.T) {
	out := directiveFor(t, "")
	if !strings.Contains(out, "validation-specialist") {
		t.Fatalf("legacy directive must name validation-specialist, got: %s", out)
	}
	if strings.Contains(out, "gatekeeper") {
		t.Fatalf("legacy directive must not demand gatekeeper evidence, got: %s", out)
	}
}

func TestHookDirectiveNamesGatekeeperForAdversarialWorkflow(t *testing.T) {
	out := directiveFor(t, workflow.ValidateContractAdversarial)
	if !strings.Contains(out, "gatekeeper") {
		t.Fatalf("adversarial directive must name gatekeeper, got: %s", out)
	}
	if strings.Contains(out, "validation-specialist") {
		t.Fatalf("adversarial directive must not demand validation-specialist, got: %s", out)
	}
	if !strings.Contains(out, "ADVERSARIAL VERIFIER") {
		t.Fatalf("validate directive must carry the delegation contract, got: %s", out)
	}
}
