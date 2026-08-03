// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: A tree that satisfies every gate completes under every profile
//
// The counterweight to every refusal test in this file set: parity must not
// be achieved by everything failing.
func TestEP_CleanTreeCompletesUnderEveryProfile(t *testing.T) {
	bin := avvBuildBin(t)
	for _, profile := range []string{"strict", "guided"} {
		dir := avvFixture(t)
		avvValidateCommands(t, dir)
		feature := "clean-" + profile
		gbdSeedWorkflow(t, dir, feature, profile)
		gbdMakeSafe(t, bin, dir, feature, profile)

		out, code := avvComplete(t, bin, dir, feature)
		if code != 0 {
			t.Fatalf("profile %q: a fully-satisfied tree must complete, got %d: %s", profile, code, out)
		}
	}
}

// Scenario: Guided relaxes process, and only process
//
// The one legitimate divergence, asserted explicitly: an identical tree
// (SAFE grounded verdict, stamped, passing suite) with NO orchestration
// evidence bundle completes under guided but is refused under strict.
func TestEP_GuidedSkipsOrchestrationEvidenceStrictRequires(t *testing.T) {
	bin := avvBuildBin(t)

	gDir := avvFixture(t)
	avvValidateCommands(t, gDir)
	gbdSeedWorkflow(t, gDir, "guided-no-evidence", "guided")
	avvWriteReport(t, gDir, "guided-no-evidence", avvReport("SAFE", "", avvGroundedCommands))
	avvStamp(t, bin, gDir, "guided-no-evidence")
	if out, code := avvComplete(t, bin, gDir, "guided-no-evidence"); code != 0 {
		t.Fatalf("guided must NOT require the orchestration evidence bundle, got %d: %s", code, out)
	}

	sDir := avvFixture(t)
	avvValidateCommands(t, sDir)
	gbdSeedWorkflow(t, sDir, "strict-no-evidence", "strict")
	avvWriteReport(t, sDir, "strict-no-evidence", avvReport("SAFE", "", avvGroundedCommands))
	avvStamp(t, bin, sDir, "strict-no-evidence")
	out, code := avvComplete(t, bin, sDir, "strict-no-evidence")
	if code == 0 {
		t.Fatalf("strict must require the orchestration evidence bundle, got exit 0: %s", out)
	}
	if !strings.Contains(strings.ToLower(out), "evidence") && !strings.Contains(out, "gatekeeper") {
		t.Fatalf("refusal must name the missing evidence: %s", out)
	}
}
