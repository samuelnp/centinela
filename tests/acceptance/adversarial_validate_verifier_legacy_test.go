// Acceptance: specs/adversarial-validate-verifier.feature
package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// Scenario: A legacy in-flight workflow still completes with an old-format
// validation-specialist report (empty ValidateContract => existence-only gate).
func TestAVV_LegacyInFlightWorkflowStillCompletes(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "already-in-flight", "") // empty => legacy
	mustWrite(t, dir+"/.workflow/already-in-flight-validation-specialist.md", "status: done\n")
	avvWriteReport(t, dir, "already-in-flight", "### Gatekeeper Report\n**Status:** SAFE\n")

	out, code := avvComplete(t, bin, dir, "already-in-flight")
	if code != 0 {
		t.Fatalf("legacy existence-only gate must still complete verbatim: %s", out)
	}
}

// Scenario: A workflow pinned to the new contract refuses a hand-authored
// legacy-format report — a fresh feature cannot dodge the new format.
func TestAVV_FreshWorkflowRefusesHandAuthoredLegacyFormat(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "fresh-workflow", "adversarial-v1")
	mustWrite(t, dir+"/.workflow/fresh-workflow-validation-specialist.md", "status: done\n")
	avvWriteReport(t, dir, "fresh-workflow", "### Gatekeeper Report\n**Status:** SAFE\n")

	out, code := avvComplete(t, bin, dir, "fresh-workflow")
	if code == 0 {
		t.Fatalf("adversarial-v1 must refuse a hand-authored legacy pair, got exit 0: %s", out)
	}
	if !strings.Contains(out, "no commands-run record") {
		t.Fatalf("message must say no commands-run record: %s", out)
	}
}

// Scenario: The validate step no longer requires validation-specialist evidence
func TestAVV_NoValidationSpecialistRequired(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "no-validation-specialist", "adversarial-v1")
	avvWriteReport(t, dir, "no-validation-specialist", avvReport("SAFE", "", avvGroundedCommands))
	avvStamp(t, bin, dir, "no-validation-specialist")

	out, code := avvComplete(t, bin, dir, "no-validation-specialist")
	if code != 0 {
		t.Fatalf("no validation-specialist evidence must not block: %s", out)
	}

	roles := orchestration.RequiredRolesForFeature("no-validation-specialist", "validate")
	if len(roles) != 1 || roles[0] != orchestration.RoleGatekeeper {
		t.Fatalf("RequiredRolesForFeature(validate) = %v, want exactly [gatekeeper]", roles)
	}
}
