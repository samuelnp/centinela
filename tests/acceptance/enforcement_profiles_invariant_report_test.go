// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: A missing verifier report blocks the validate step under every profile
func TestEP_MissingGatekeeperReportBlocksEveryProfile(t *testing.T) {
	bin := avvBuildBin(t)
	for _, profile := range []string{"strict", "guided"} {
		dir := avvFixture(t)
		avvValidateCommands(t, dir)
		feature := "missing-report-" + profile
		gbdSeedWorkflow(t, dir, feature, profile)
		// Deliberately NO gatekeeper report written at all.

		out, code := avvComplete(t, bin, dir, feature)
		if code == 0 {
			t.Fatalf("profile %q: a missing gatekeeper report must block completion, got exit 0: %s", profile, out)
		}
		if !strings.Contains(out, "gatekeeper report not found") {
			t.Fatalf("profile %q: message must name the missing report: %s", profile, out)
		}
	}
}

// Scenario: An ungrounded verifier verdict blocks the validate step under every profile
func TestEP_UngroundedVerdictBlocksEveryProfile(t *testing.T) {
	bin := avvBuildBin(t)
	for _, profile := range []string{"strict", "guided"} {
		dir := avvFixture(t)
		avvValidateCommands(t, dir)
		feature := "ungrounded-" + profile
		gbdSeedWorkflow(t, dir, feature, profile)
		// A narrated SAFE verdict with an empty commands array — no
		// commands-run record and no verified revision.
		avvWriteReport(t, dir, feature, avvReport("SAFE", "", "[]"))

		out, code := avvComplete(t, bin, dir, feature)
		if code == 0 {
			t.Fatalf("profile %q: an ungrounded verdict must block completion, got exit 0: %s", profile, out)
		}
		if !strings.Contains(out, "no commands-run record") {
			t.Fatalf("profile %q: message must say no commands-run record: %s", profile, out)
		}
	}
}
