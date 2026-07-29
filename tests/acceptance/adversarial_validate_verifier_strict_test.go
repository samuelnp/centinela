// Acceptance: specs/adversarial-validate-verifier.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: A legacy in-flight workflow still completes with an old-format
// validation-specialist report — in STRICT mode.
//
// Regression: the hook resolved required roles through the contract-blind
// policy layer while `complete` resolved them through the pinned contract, so
// a legacy strict workflow was told to write gatekeeper evidence and then
// hard-blocked for missing validation-specialist evidence. The directive is
// the contract: whatever it names, the gate must accept.
func TestAVV_StrictLegacyDirectiveMatchesGate(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedStrict(t, dir, "legacy-strict", "") // empty => legacy
	avvValidateCommands(t, dir)

	directive := avvDirective(t, bin, dir)
	if !strings.Contains(directive, "validation-specialist") {
		t.Fatalf("legacy directive must name validation-specialist: %s", directive)
	}
	if strings.Contains(directive, "gatekeeper.json") {
		t.Fatalf("legacy directive must not demand gatekeeper evidence: %s", directive)
	}

	avvWriteEvidence(t, bin, dir, "legacy-strict", avvRequiredEvidence(t, directive, "legacy-strict"))
	avvWriteReport(t, dir, "legacy-strict", "### Gatekeeper Report\n**Status:** SAFE\n")
	if out, code := avvComplete(t, bin, dir, "legacy-strict"); code != 0 {
		t.Fatalf("the gate refused the evidence its own directive demanded: %s", out)
	}
}

// Scenario: The validate step no longer requires validation-specialist
// evidence — in STRICT mode, where evidence JSON is actually enforced.
func TestAVV_StrictAdversarialDirectiveMatchesGate(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedStrict(t, dir, "strict-adversarial", "adversarial-v1")
	avvValidateCommands(t, dir)

	directive := avvDirective(t, bin, dir)
	if !strings.Contains(directive, "gatekeeper") {
		t.Fatalf("adversarial directive must name gatekeeper: %s", directive)
	}
	if strings.Contains(directive, "validation-specialist") {
		t.Fatalf("adversarial directive must not demand validation-specialist: %s", directive)
	}

	avvWriteEvidence(t, bin, dir, "strict-adversarial", avvRequiredEvidence(t, directive, "strict-adversarial"))
	avvWriteReport(t, dir, "strict-adversarial", avvReport("SAFE", "", avvGroundedCommands))
	avvStamp(t, bin, dir, "strict-adversarial") // stamping is always the LAST action
	if out, code := avvComplete(t, bin, dir, "strict-adversarial"); code != 0 {
		t.Fatalf("the gate refused the evidence its own directive demanded: %s", out)
	}
}

// Scenario: a strict adversarial workflow that produced only the DEPRECATED
// role's evidence is still refused — the contract cannot be dodged.
func TestAVV_StrictAdversarialRefusesLegacyEvidenceOnly(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedStrict(t, dir, "dodge-attempt", "adversarial-v1")
	avvValidateCommands(t, dir)
	avvWriteEvidence(t, bin, dir, "dodge-attempt", []string{
		".workflow/dodge-attempt-validation-specialist.json",
	})
	avvWriteReport(t, dir, "dodge-attempt", avvReport("SAFE", "", avvGroundedCommands))
	avvStamp(t, bin, dir, "dodge-attempt")

	out, code := avvComplete(t, bin, dir, "dodge-attempt")
	if code == 0 {
		t.Fatalf("legacy evidence must not satisfy an adversarial-v1 workflow: %s", out)
	}
	if !strings.Contains(out, "gatekeeper") {
		t.Fatalf("block must name the gatekeeper evidence it wants: %s", out)
	}
}
