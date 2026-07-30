// Acceptance: specs/truthful-validators.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: The reused single-run suite outcome is analysed for acceptance skips
//
// The PriorTestRun reuse path only fires from `centinela complete` at the
// validate step (cmd/centinela/complete_verify.go): it needs a real git tree,
// workflow.VerificationFresh, and a genuine executeValidation() run, so it is
// not practical to drive end-to-end here — the SAME tradeoff this repo already
// made for this exact class of assembly wiring (see
// tests/acceptance/dedupe_validate_suite_runs_complete_test.go, which pins the
// reuse ordering by source rather than a live run). The behavioral proof that
// a reused outcome is judged for skips is
// internal/verify/claim_tests_acceptance_test.go's
// TestCheckTestsPass_PriorRunAppliesTheAcceptanceRule (real assertions against
// checkTestsPass, not a stub). This test pins the two wiring points that
// assertion depends on staying connected: the outcome is threaded into
// deps.PriorTestRun, and checkTestsPass classifies it over the WHOLE
// configured command set (not a real command string, so per-command
// classification cannot apply to it).
func TestTV_Verify_PriorRunWiringReachesTheAcceptanceRule(t *testing.T) {
	completeVerify := tvSource(t, "cmd", "centinela", "complete_verify.go")
	if !strings.Contains(completeVerify, "deps.PriorTestRun = prior") {
		t.Fatalf("runClaimVerification must thread the gate's own run into deps.PriorTestRun:\n%s", completeVerify)
	}
	claimTests := tvSource(t, "internal", "verify", "claim_tests.go")
	if !strings.Contains(claimTests, "acceptance.AnyExecutionCommand(cfg.Validate.Commands)") {
		t.Fatalf("checkTestsPass must classify a PriorTestRun over the configured command set:\n%s", claimTests)
	}
	if !strings.Contains(claimTests, "if deps.PriorTestRun != nil {") {
		t.Fatalf("checkTestsPass must branch on PriorTestRun before running commands itself:\n%s", claimTests)
	}
}

// tvSource reads a source file under the repo root and fails the test if it
// cannot be read.
func tvSource(t *testing.T, parts ...string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(append([]string{repoRoot(t)}, parts...)...))
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	return string(data)
}
