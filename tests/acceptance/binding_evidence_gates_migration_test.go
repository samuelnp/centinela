// Acceptance: specs/binding-evidence-gates.feature
//
// The two migration behaviours: what the alternate-pin tolerance refuses, and
// what evidence seeded by the old hardcoded prefill costs to repair.
package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Scenario: The tolerance refuses a role from another step
//
// The tolerance exists so evidence naming the successor STEP under the other
// contract pin keeps completing. It is scoped to that one step: a role from
// any other step is refused, so it can never be used to skip a step.
func TestBEG_ToleranceRefusesAnotherStepsRole(t *testing.T) {
	for _, got := range []string{"documentation-specialist", "senior-engineer", "planner", "complete"} {
		if err := begTestsStep(t, "demo", got, false); err == nil {
			t.Errorf("handoffTo %q belongs to another step and must be refused", got)
		}
	}
	for _, got := range []string{"gatekeeper", "validation-specialist"} {
		if err := begTestsStep(t, "demo", got, false); err != nil {
			t.Errorf("handoffTo %q names the successor step and must be accepted: %v", got, err)
		}
	}
}

// Scenario: Evidence seeded by the old prefill on a user-facing feature
//
// The decision: NOT accepted. The old prefill's "qa-senior" does not misname
// the successor step's occupant — it names a LATER step's role, asserting that
// the same-step ux-ui-specialist hop does not exist. Widening the tolerance to
// cover it would give back the one property that makes the gate meaningful,
// that a required same-step role cannot be skipped. The migration cost is the
// single command the error already prints, and this test proves that running
// it makes the same completion succeed.
func TestBEG_OldPrefillOnUserFacingFeatureIsRefusedThenRepairable(t *testing.T) {
	begSave(t, begRepo(t, "demo-ux", true))
	begEvidence(t, "demo-ux", "code", orchestration.RoleSeniorEngineer, "qa-senior")
	begEvidence(t, "demo-ux", "code", orchestration.RoleUXUISpecialist, "qa-senior")

	err := workflow.ValidateArtifacts("demo-ux", "code", begCfg())
	if err == nil {
		t.Fatal("a same-step hop is exact: the pre-gate literal must be refused")
	}
	remedy := "centinela evidence set demo-ux senior-engineer handoffTo ux-ui-specialist"
	if !strings.Contains(err.Error(), remedy) {
		t.Fatalf("the refusal must print the runnable remedy, got %v", err)
	}
	// Apply exactly what the error said, and the same gate now passes.
	begEvidence(t, "demo-ux", "code", orchestration.RoleSeniorEngineer, "ux-ui-specialist")
	if err := workflow.ValidateArtifacts("demo-ux", "code", begCfg()); err != nil {
		t.Fatalf("the printed remedy must actually unblock completion: %v", err)
	}
}
