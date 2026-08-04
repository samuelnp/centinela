// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: Guided relaxes process, and only process
//
// This test is the "same verification freshness checks" clause of that
// scenario: a revision-skewed gatekeeper stamp must demand a fresh verifier
// identically under both profiles, exactly like TestAVV_RevisionSkewDemandsFreshVerification
// already proves for a profile-less legacy fixture.
func TestEP_StaleVerificationBlocksEveryProfile(t *testing.T) {
	bin := avvBuildBin(t)
	for _, profile := range []string{"strict", "guided"} {
		dir := avvFixture(t)
		avvValidateCommands(t, dir)
		feature := "stale-" + profile
		gbdSeedWorkflow(t, dir, feature, profile)
		gbdMakeSafe(t, bin, dir, feature, profile)

		mustWrite(t, filepath.Join(dir, "src.go"), "package x\n\nfunc fix() {}\n")
		commit(t, dir, "fix landed on top of the verified commit")

		out, code := avvComplete(t, bin, dir, feature)
		if code == 0 {
			t.Fatalf("profile %q: revision skew must block completion, got exit 0: %s", profile, out)
		}
		if !strings.Contains(out, "stale") || !strings.Contains(out, "FRESH verifier") {
			t.Fatalf("profile %q: message must say stale and demand a fresh verifier: %s", profile, out)
		}
	}
}
