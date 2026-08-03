// Acceptance: specs/binding-evidence-gates.feature
package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// begSteward writes merge-steward evidence carrying handoffTo and runs the
// orchestration evidence validator directly — the steward is out-of-band, so
// it is never reached through a workflow step gate.
func begSteward(t *testing.T, handoffTo string) error {
	t.Helper()
	begSave(t, begRepo(t, "demo", false))
	begEvidence(t, "demo", "merge", orchestration.RoleMergeSteward, handoffTo)
	path := orchestration.JSONPath("demo", orchestration.RoleMergeSteward)
	return orchestration.ValidateEvidence(path, "demo", "merge",
		orchestration.RoleMergeSteward, config.UIPaths(begCfg()))
}

// Scenario: A rejected merge-steward handoff outside {complete, user}
func TestBEG_StewardHandoffOutsideVerdictPairRejected(t *testing.T) {
	err := begSteward(t, "orchestrator")
	if err == nil {
		t.Fatal("a steward handoff outside the verdict pair must be refused")
	}
	for _, want := range []string{"complete", "user", "orchestrator"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error must name %q: %v", want, err)
		}
	}
}

// Scenario: A valid merge-steward escalation handoff
func TestBEG_StewardEscalationHandoffAccepted(t *testing.T) {
	if err := begSteward(t, "user"); err != nil {
		t.Fatalf("ESCALATE is a legal steward verdict: %v", err)
	}
	if err := begSteward(t, "complete"); err != nil {
		t.Fatalf("APPLY is a legal steward verdict: %v", err)
	}
}
