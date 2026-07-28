// Acceptance: specs/adversarial-validate-verifier.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: A grounded SAFE verdict with a matching revision advances validate
func TestAVV_SafeVerdictAdvances(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "checkout-fix", "adversarial-v1")
	avvWriteReport(t, dir, "checkout-fix", avvReport("SAFE", "", avvGroundedCommands))
	avvStamp(t, bin, dir, "checkout-fix")

	out, code := avvComplete(t, bin, dir, "checkout-fix")
	if code != 0 {
		t.Fatalf("SAFE verdict must advance, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Next step: docs") {
		t.Fatalf("docs step should become active: %s", out)
	}
}

// Scenario: A WARNING verdict advances validate but the finding is recorded
func TestAVV_WarningVerdictAdvancesAndLandsInLedger(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "rate-limit-tweak", "adversarial-v1")
	avvWriteReport(t, dir, "rate-limit-tweak",
		avvReport("WARNING", "- retry budget could starve under burst load", avvGroundedCommands))
	avvStamp(t, bin, dir, "rate-limit-tweak")

	out, code := avvComplete(t, bin, dir, "rate-limit-tweak")
	if code != 0 {
		t.Fatalf("WARNING verdict must advance, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Next step: docs") {
		t.Fatalf("docs step should become active: %s", out)
	}

	entries, err := os.ReadDir(filepath.Join(dir, ".workflow", "memory", "entries"))
	if err != nil || len(entries) == 0 {
		t.Fatalf("WARNING finding must land in the memory ledger: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(dir, ".workflow", "memory", "entries", entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "retry budget could starve under burst load") {
		t.Fatalf("ledger entry must carry the finding text: %s", body)
	}
}
