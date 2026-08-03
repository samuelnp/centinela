// Acceptance: specs/guided-by-default.feature
//
// Shared fixtures for the guided-by-default proof-parity table, which extends
// enforcement_profiles_invariant_test.go rather than standing up a second
// binary-building/fixture mechanism. Every helper here reuses the adversarial
// verifier helpers (avvBuildBin, avvFixture, avvWriteReport, avvStamp, ...)
// defined in adversarial_validate_verifier_helper_test.go and
// adversarial_validate_verifier_strict_helper_test.go.
package acceptance_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// gbdOrchestrationMode mirrors what a real NewWithOrder pins for profile, so
// the hand-written fixture matches production state shape exactly.
func gbdOrchestrationMode(profile string) string {
	if profile == "strict" {
		return workflow.StrictOrchestrationMode
	}
	return ""
}

// gbdSeedWorkflow writes a workflow state at the validate step, pinned (tier
// 1) to profile via an explicit per-feature EnforcementProfile so
// EffectiveProfile resolves deterministically regardless of any config in
// dir. validateContract is always adversarial-v1 — proof parity must hold
// under the CURRENT gate, never the legacy one.
func gbdSeedWorkflow(t *testing.T, dir, feature, profile string) {
	t.Helper()
	mode := gbdOrchestrationMode(profile)
	body := fmt.Sprintf(`{"feature":%q,"currentStep":"validate",`+
		`"stepOrder":["plan","code","tests","validate","docs"],"steps":{},`+
		`"validateContract":"adversarial-v1","enforcementProfile":%q,"orchestrationMode":%q}`,
		feature, profile, mode)
	mustWrite(t, filepath.Join(dir, ".workflow", feature+".json"), body)
}

// gbdMakeSafe drives feature to a grounded SAFE verdict. Under strict it also
// produces the orchestration-evidence bundle the directive demands, reusing
// the exact avv helpers the strict-mode fixtures already rely on; under
// guided no evidence is ever written, matching production (orchestrationMode
// empty short-circuits validateOrchestration).
func gbdMakeSafe(t *testing.T, bin, dir, feature, profile string) {
	t.Helper()
	if profile == "strict" {
		directive := avvDirective(t, bin, dir)
		avvWriteEvidence(t, bin, dir, feature, avvRequiredEvidence(t, directive, feature))
	}
	avvWriteReport(t, dir, feature, avvReport("SAFE", "", avvGroundedCommands))
	avvStamp(t, bin, dir, feature)
}
