package workflow

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func seedContract(t *testing.T, contract string) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	wf := New("f")
	wf.ValidateContract = contract
	if err := Save(wf); err != nil {
		t.Fatal(err)
	}
}

// The directive and the gate MUST resolve the same roles: a contract-blind
// resolver told a legacy workflow to write gatekeeper evidence while the gate
// demanded validation-specialist.
func TestRequiredEvidenceRolesLegacyValidate(t *testing.T) {
	seedContract(t, "")
	roles := RequiredEvidenceRoles("f", "validate")
	if len(roles) != 1 || roles[0] != orchestration.RoleValidationSpec {
		t.Fatalf("legacy validate roles = %v, want [validation-specialist]", roles)
	}
}

func TestRequiredEvidenceRolesAdversarialValidate(t *testing.T) {
	seedContract(t, ValidateContractAdversarial)
	roles := RequiredEvidenceRoles("f", "validate")
	if len(roles) != 1 || roles[0] != orchestration.RoleGatekeeper {
		t.Fatalf("adversarial validate roles = %v, want [gatekeeper]", roles)
	}
}

// Non-validate steps are contract-independent and defer to policy.
func TestRequiredEvidenceRolesOtherStepsDeferToPolicy(t *testing.T) {
	seedContract(t, "")
	for _, step := range []string{"plan", "code", "tests", "docs"} {
		got := RequiredEvidenceRoles("f", step)
		want := orchestration.RequiredRolesForFeature("f", step)
		if len(got) != len(want) {
			t.Fatalf("%s roles = %v, want %v", step, got, want)
		}
	}
}
