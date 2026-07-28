// Acceptance: specs/adversarial-validate-verifier.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: A CRITICAL verdict blocks complete with the finding echoed
func TestAVV_CriticalVerdictBlocksWithFindingEchoed(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "leaky-session", "adversarial-v1")
	avvWriteReport(t, dir, "leaky-session",
		avvReport("CRITICAL", "- session tokens are never invalidated on logout", avvGroundedCommands))

	out, code := avvComplete(t, bin, dir, "leaky-session")
	if code == 0 {
		t.Fatalf("CRITICAL verdict must hard-block, got exit 0: %s", out)
	}
	if !strings.Contains(out, "session tokens are never invalidated on logout") {
		t.Fatalf("block message must echo the finding: %s", out)
	}
	if !strings.Contains(out, "FRESH verifier") {
		t.Fatalf("block message must instruct fresh re-verification: %s", out)
	}
}

// Scenario Outline: Legacy severity aliases normalize to CRITICAL and block complete
func TestAVV_LegacyAliasesBlockAsCritical(t *testing.T) {
	bin := avvBuildBin(t)
	for _, tc := range []struct{ feature, alias string }{
		{"legacy-blocking", "BLOCKING"},
		{"legacy-unsafe", "UNSAFE"},
	} {
		t.Run(tc.alias, func(t *testing.T) {
			dir := avvFixture(t)
			avvSeedWorkflow(t, dir, tc.feature, "adversarial-v1")
			avvWriteReport(t, dir, tc.feature, avvReport(tc.alias, "- finding", avvGroundedCommands))
			out, code := avvComplete(t, bin, dir, tc.feature)
			if code == 0 {
				t.Fatalf("%s alias must block as CRITICAL, got exit 0: %s", tc.alias, out)
			}
			if !strings.Contains(out, "CRITICAL") {
				t.Fatalf("%s alias must normalize to CRITICAL in the message: %s", tc.alias, out)
			}
		})
	}
}

// Scenario: A missing Status line blocks complete
func TestAVV_MissingStatusLineBlocks(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "no-status-line", "adversarial-v1")
	avvWriteReport(t, dir, "no-status-line", "### Report\n\n#### Findings\n- something\n")

	out, code := avvComplete(t, bin, dir, "no-status-line")
	if code == 0 {
		t.Fatalf("missing Status must block, got exit 0: %s", out)
	}
	if !strings.Contains(out, "missing or unparseable") {
		t.Fatalf("message must say the verdict is missing or unparseable: %s", out)
	}
}

// Scenario: An unparseable Status line blocks complete, never fails open
func TestAVV_UnparseableStatusBlocks(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "garbled-status", "adversarial-v1")
	avvWriteReport(t, dir, "garbled-status", "### Report\n**Status:** mostly fine I think\n")

	out, code := avvComplete(t, bin, dir, "garbled-status")
	if code == 0 {
		t.Fatalf("unparseable Status must block, got exit 0: %s", out)
	}
	if !strings.Contains(out, "missing or unparseable") {
		t.Fatalf("message must say the verdict is missing or unparseable: %s", out)
	}
}
