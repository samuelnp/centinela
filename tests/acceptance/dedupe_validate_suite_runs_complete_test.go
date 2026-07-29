package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/dedupe-validate-suite-runs.feature
//
// The complete-gate wiring is asserted against the cmd source: the reuse
// is only sound because of the exact ordering executeValidation → second
// VerificationFresh → runClaimVerification(..., completedValidationOutcome()).

func cmdSource(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "cmd", "centinela", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}

// Scenario: complete's claim verification reuses the gate's own run
func TestCompleteClaimVerificationReusesGatesOwnRun(t *testing.T) {
	gates := cmdSource(t, "complete_validate_gates.go")
	if !strings.Contains(gates, "runClaimVerification(feature, step, model, cfg, completedValidationOutcome())") {
		t.Fatal("runValidateGates must pass the gate's own synthesized outcome to claim verification")
	}
	verifySrc := cmdSource(t, "complete_verify.go")
	if !strings.Contains(verifySrc, "deps.PriorTestRun = prior") {
		t.Fatal("runClaimVerification must wire the prior outcome into deps.PriorTestRun")
	}
}

// Scenario: a failing validate command blocks completion before any reuse
func TestFailingValidateBlocksCompletionBeforeReuse(t *testing.T) {
	gates := cmdSource(t, "complete_validate_gates.go")
	failCheck := strings.Index(gates, "if err := executeValidation(); err != nil")
	reuse := strings.Index(gates, "completedValidationOutcome()")
	if failCheck < 0 {
		t.Fatal("executeValidation failure must return early")
	}
	if reuse < 0 || reuse < failCheck {
		t.Fatal("the outcome must only be synthesized after executeValidation succeeded")
	}
}

// Scenario: a tree change during the gate's suite run blocks completion
func TestTreeChangeDuringSuiteRunBlocksCompletion(t *testing.T) {
	gates := cmdSource(t, "complete_validate_gates.go")
	run := strings.Index(gates, "executeValidation()")
	reuse := strings.Index(gates, "completedValidationOutcome()")
	recheck := strings.Index(gates[run:], "workflow.VerificationFresh")
	if run < 0 || reuse < 0 || recheck < 0 {
		t.Fatal("expected executeValidation, a post-run VerificationFresh re-check, and the reuse call")
	}
	if run+recheck > reuse {
		t.Fatal("VerificationFresh must be re-checked AFTER the suite run and BEFORE any reuse")
	}
}

// Scenario: standalone verify surfaces still re-run the suite for real
func TestStandaloneVerifySurfacesStillRerunSuite(t *testing.T) {
	deps := cmdSource(t, "verify_deps.go")
	if strings.Contains(deps, "PriorTestRun") {
		t.Fatal("verifyDepsFor must never set PriorTestRun — verify/verdict/MCP re-run for real")
	}
	verifySrc := cmdSource(t, "complete_verify.go")
	if !strings.Contains(verifySrc, "deps := verifyDepsFor(feature, step)") {
		t.Fatal("complete's claim verification must still build deps via verifyDepsFor")
	}
}
