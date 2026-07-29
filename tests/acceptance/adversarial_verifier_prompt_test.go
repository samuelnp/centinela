// Acceptance: specs/adversarial-validate-verifier.feature
//
// Asserts docs/architecture/gatekeeper-prompt.md carries the refutation-stance
// rewrite (plan Slice 5 / brief AC1). Generic CLI-mandate and Deferred-Findings
// block coverage lives in prompts_mandate_cli_acceptance_test.go and
// deferred_findings_prompt_parity_test.go — this file is feature-specific.
package acceptance_test

import (
	"os"
	"strings"
	"testing"
)

func TestAVVPrompt_RefutationStanceAndGroundingContract(t *testing.T) {
	data, err := os.ReadFile("../../docs/architecture/gatekeeper-prompt.md")
	if err != nil {
		t.Fatal(err)
	}
	// Collapse whitespace so a substring is never split by the source's
	// prose line-wrapping (e.g. "Never\nnarrate a pass.").
	body := strings.Join(strings.Fields(string(data)), " ")

	for _, want := range []string{
		// Refutation stance (AC1).
		"is FALSE",
		"not auditing compliance",
		"Default to NOT-VERIFIED",
		// Paths-only input contract (AC1).
		"Input Contract",
		"diff versus the base ref",
		"specs/<FEATURE_NAME>.feature",
		// No-orchestrator-summaries prohibition (AC1, edge case).
		"Do NOT accept the orchestrator's summary",
		"contaminated delegation",
		// Verdict vocabulary (D5).
		"CRITICAL",
		// The machine-readable grounding fence (D2).
		"```json centinela:verification",
		"\"revision\"",
		"\"treeDigest\"",
		"\"commands\"",
		// Mandatory stamp instruction (AC5).
		"centinela artifact stamp <FEATURE_NAME>",
		// Fail-closed clause (edge case).
		"Fail-Closed Clause",
		"leave the `commands` array empty",
		"Never narrate a pass",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("gatekeeper-prompt.md missing %q", want)
		}
	}
}
