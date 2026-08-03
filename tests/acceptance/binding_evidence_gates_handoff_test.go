// Acceptance: specs/binding-evidence-gates.feature
package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// begTestsStep prepares a tests-step feature with qa-senior evidence carrying
// handoffTo, and returns the artifact gate's verdict.
func begTestsStep(t *testing.T, feature, handoffTo string, legacyValidate bool) error {
	t.Helper()
	wf := begRepo(t, feature, false)
	if legacyValidate {
		wf.ValidateContract = ""
	}
	begSave(t, wf)
	begTestsStepArtifacts(t, feature)
	begEvidence(t, feature, "tests", orchestration.RoleQASeniorEngineer, handoffTo)
	return workflow.ValidateArtifacts(feature, "tests", begCfg())
}

// Scenario: A rejected banana handoff
func TestBEG_RejectedBananaHandoff(t *testing.T) {
	err := begTestsStep(t, "demo", "banana", false)
	if err == nil {
		t.Fatal("an out-of-chain handoffTo must block completion")
	}
	// Derived from the workflow's own contract, not a hardcoded literal: this
	// workflow pins adversarial-v1, so its validate step is the gatekeeper's.
	for _, want := range []string{`"banana"`, "gatekeeper", "centinela evidence set demo qa-senior handoffTo gatekeeper"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error must name %q, got %v", want, err)
		}
	}
}

// Scenario: A valid terminal handoff on an internal feature
func TestBEG_TerminalHandoffOnInternalFeature(t *testing.T) {
	wf := begRepo(t, "demo-internal", false)
	begSave(t, wf)
	// An internal feature's docs step requires no role evidence, so validate is
	// terminal — no hardcoded "documentation-specialist" is involved.
	if roles := workflow.RequiredEvidenceRoles("demo-internal", "docs"); len(roles) != 0 {
		t.Fatalf("internal docs step must require no role evidence, got %v", roles)
	}
	begEvidence(t, "demo-internal", "validate", orchestration.RoleGatekeeper, "complete")
	begGroundedReport(t, "demo-internal")
	if err := workflow.ValidateArtifacts("demo-internal", "validate", begCfg()); err != nil {
		t.Fatalf("a terminal handoff must complete: %v", err)
	}
}

// Scenario: A valid mid-chain handoff on a legacy workflow
func TestBEG_MidChainHandoffOnLegacyWorkflow(t *testing.T) {
	if err := begTestsStep(t, "demo-legacy", "validation-specialist", true); err != nil {
		t.Fatalf("a legacy pin's own successor must be in-chain: %v", err)
	}
	// ...and the literal a FRESH workflow would require is still accepted, so
	// evidence seeded before this gate keeps completing either way.
	if err := begTestsStep(t, "demo-legacy", "gatekeeper", true); err != nil {
		t.Fatalf("the successor step's other occupant must stay in-chain: %v", err)
	}
}

// Scenario: A valid same-step handoff when UX is required
func TestBEG_SameStepHandoffWhenUXRequired(t *testing.T) {
	begSave(t, begRepo(t, "demo-ux", true))
	begEvidence(t, "demo-ux", "code", orchestration.RoleSeniorEngineer, "ux-ui-specialist")
	begEvidence(t, "demo-ux", "code", orchestration.RoleUXUISpecialist, "qa-senior")
	if err := workflow.ValidateArtifacts("demo-ux", "code", begCfg()); err != nil {
		t.Fatalf("the next required role WITHIN the step must be in-chain: %v", err)
	}
}
