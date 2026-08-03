// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: A blocking production-readiness report blocks completion under every profile
// Scenario: The production-readiness gate is untouched by cascade slimming
func TestEP_BlockingProductionReadinessBlocksEveryProfile(t *testing.T) {
	bin := avvBuildBin(t)
	for _, profile := range []string{"strict", "guided"} {
		dir := avvFixture(t)
		mustWrite(t, filepath.Join(dir, "centinela.toml"),
			"[validate]\ncommands = [\"true\"]\n[gates]\nproduction_readiness = true\n")
		feature := "pr-blocking-" + profile
		gbdSeedWorkflow(t, dir, feature, profile)
		gbdMakeSafe(t, bin, dir, feature, profile)
		mustWrite(t, filepath.Join(dir, ".workflow", feature+"-production-readiness.md"),
			"### Production Readiness Report\n**Status:** BLOCKING\n")

		out, code := avvComplete(t, bin, dir, feature)
		if code == 0 {
			t.Fatalf("profile %q: BLOCKING production readiness must refuse completion, got exit 0: %s", profile, out)
		}
		if !strings.Contains(out, "BLOCKING") {
			t.Fatalf("profile %q: message must cite the BLOCKING status: %s", profile, out)
		}
	}
}
