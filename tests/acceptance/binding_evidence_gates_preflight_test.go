// Acceptance: specs/binding-evidence-gates.feature
//
// The pre-flight/gate agreement (edge-case finding EC-7). No spec scenario
// yet: the prewrite hook blocks specs/ edits during the tests step.
package acceptance_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// begSeedTestsStep writes a strict tests-step workflow into dir.
func begSeedTestsStep(t *testing.T, dir, feature string) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, "docs/features", feature+".md"), "# "+feature+"\n")
	mustWrite(t, filepath.Join(dir, ".workflow", feature+".json"), fmt.Sprintf(
		`{"feature":%q,"currentStep":"tests","stepOrder":["plan","code","tests","validate","docs"],`+
			`"steps":{},"validateContract":"adversarial-v1","planContract":"planner-v1",`+
			`"orchestrationMode":"strict-subagents-v1"}`, feature))
}

// TestBEG_EvidenceValidateAgreesWithComplete closes the asymmetry that made
// the agents' own self-check a false-success generator: `evidence set` accepted
// an out-of-chain handoffTo, `evidence validate` then printed "evidence ok",
// and `complete` refused it. The pre-flight now asks the gate's own question.
func TestBEG_EvidenceValidateAgreesWithComplete(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	begSeedTestsStep(t, dir, "demo")
	if out, code := runCent(t, bin, dir, "evidence", "init", "demo", "qa-senior"); code != 0 {
		t.Fatalf("evidence init must succeed: %s", out)
	}
	if out, code := runCent(t, bin, dir, "evidence", "set", "demo", "qa-senior", "handoffTo", "banana"); code != 0 {
		t.Fatalf("evidence set must succeed: %s", out)
	}
	out, code := runCent(t, bin, dir, "evidence", "validate", "demo")
	if code == 0 || strings.Contains(out, "evidence ok") {
		t.Fatalf("the pre-flight must not green-light what complete refuses: (%d) %s", code, out)
	}
	if !strings.Contains(out, "handoffTo") || !strings.Contains(out, "gatekeeper") {
		t.Fatalf("the pre-flight must name the derived successor: %s", out)
	}
	// Applying the derived successor makes the same pre-flight agree again.
	if out, code := runCent(t, bin, dir, "evidence", "set", "demo", "qa-senior", "handoffTo", "gatekeeper"); code != 0 {
		t.Fatalf("evidence set must succeed: %s", out)
	}
	// (Other fields are still unfilled here — only the CHAIN complaint must go.)
	if out, code := runCent(t, bin, dir, "evidence", "validate", "demo"); strings.Contains(out, "its successor") {
		t.Fatalf("an in-chain handoffTo must not be reported (%d): %s", code, out)
	}
}
