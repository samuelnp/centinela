package workflow

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// acceptedHandoffSet returns every candidate the chain check admits for a
// next-step hop out of the tests step under the given contract pin.
func acceptedHandoffSet(t *testing.T, legacyValidate bool, candidates []string) map[string]bool {
	t.Helper()
	accepted := map[string]bool{}
	for _, got := range candidates {
		tc := hoCase{legacyValidate: legacyValidate, step: "tests", role: orchestration.RoleQASeniorEngineer}
		t.Run(got, func(t *testing.T) {
			if hoChain(t, tc, got) == nil {
				accepted[got] = true
			}
		})
	}
	return accepted
}

// TestContractPinFlipDoesNotWidenAcceptance is the single property that makes
// the tolerance auditable: flipping which contract a workflow pins swaps which
// name is CANONICAL, never how many names are legal. Without it, a future
// widening of alternateContractRoles would fail no test.
func TestContractPinFlipDoesNotWidenAcceptance(t *testing.T) {
	candidates := append([]string{"gatekeeper", "validation-specialist"}, outOfChainValues...)
	modern := acceptedHandoffSet(t, false, candidates)
	legacy := acceptedHandoffSet(t, true, candidates)
	want := map[string]bool{"gatekeeper": true, "validation-specialist": true}
	for _, set := range []map[string]bool{modern, legacy} {
		if len(set) != len(want) {
			t.Fatalf("accepted set %v, want %v", set, want)
		}
		for k := range want {
			if !set[k] {
				t.Fatalf("accepted set %v missing %q", set, k)
			}
		}
	}
}

// Unreadable or unset evidence is SILENT here: orchestration.ValidateRoles has
// already named it, and a second message for one missing file buries the
// remedy. Pinned so the no-double-report contract cannot regress into noise.
func TestHandoffChainSilentOnUnreadableEvidence(t *testing.T) {
	tc := hoCase{step: "tests", role: orchestration.RoleQASeniorEngineer}
	roles := []orchestration.Role{orchestration.RoleQASeniorEngineer}

	t.Run("missing file", func(t *testing.T) {
		hoRepo(t, "f", tc)
		if err := validateHandoffChain("f", "tests", roles); err != nil {
			t.Fatalf("missing evidence must be silent here, got %v", err)
		}
	})
	t.Run("unparseable file", func(t *testing.T) {
		hoRepo(t, "f", tc)
		if err := os.WriteFile(orchestration.JSONPath("f", orchestration.RoleQASeniorEngineer), []byte("{nope"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := validateHandoffChain("f", "tests", roles); err != nil {
			t.Fatalf("unparseable evidence must be silent here, got %v", err)
		}
	})
	t.Run("empty handoffTo", func(t *testing.T) {
		hoRepo(t, "f", tc)
		hoEvidence(t, "f", orchestration.RoleQASeniorEngineer, "")
		if err := validateHandoffChain("f", "tests", roles); err != nil {
			t.Fatalf("an empty handoffTo is ValidateEvidence's rule, got %v", err)
		}
	})
	t.Run("absent workflow state", func(t *testing.T) {
		t.Chdir(t.TempDir())
		if err := validateHandoffChain("ghost", "tests", roles); err != nil {
			t.Fatalf("no contract means nothing to check, got %v", err)
		}
	})
}
