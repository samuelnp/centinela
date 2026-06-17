package acceptance_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Init pre-fills plan-snapshot inputs for big-thinker
func TestDAS_InitPreFillsBigThinker(t *testing.T) {
	dasChdir(t, "demo.md", "other.md")
	dasInit(t, "demo", orchestration.RoleBigThinker)
	_, e := dasReadJSON(t, "demo", orchestration.RoleBigThinker)
	if !reflect.DeepEqual(e.Inputs, orchestration.RequiredPlanInputs("demo")) {
		t.Fatalf("inputs != RequiredPlanInputs: %v", e.Inputs)
	}
	for _, want := range []string{"docs/features/demo.md", "docs/features/other.md", "docs/plans/demo.md"} {
		if !dasContains(e.Inputs, want) {
			t.Fatalf("inputs missing %q: %v", want, e.Inputs)
		}
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Init pre-fills plan-snapshot inputs for feature-specialist
func TestDAS_InitPreFillsFeatureSpecialist(t *testing.T) {
	dasChdir(t, "demo.md")
	dasInit(t, "demo", orchestration.RoleFeatureSpecial)
	_, e := dasReadJSON(t, "demo", orchestration.RoleFeatureSpecial)
	if !reflect.DeepEqual(e.Inputs, orchestration.RequiredPlanInputs("demo")) {
		t.Fatalf("inputs != RequiredPlanInputs: %v", e.Inputs)
	}
	if !dasContains(e.Inputs, "docs/plans/demo.md") {
		t.Fatalf("inputs missing plan path: %v", e.Inputs)
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Init leaves inputs empty for senior-engineer
func TestDAS_InitInputsEmptySeniorEngineer(t *testing.T) {
	dasChdir(t, "demo.md")
	dasInit(t, "demo", orchestration.RoleSeniorEngineer)
	_, e := dasReadJSON(t, "demo", orchestration.RoleSeniorEngineer)
	if len(e.Inputs) != 0 {
		t.Fatalf("senior-engineer inputs not empty: %v", e.Inputs)
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Init leaves inputs empty for every non-plan role
func TestDAS_InitInputsEmptyEveryNonPlanRole(t *testing.T) {
	roles := []evidence.Role{
		orchestration.RoleSeniorEngineer, orchestration.RoleUXUISpecialist,
		orchestration.RoleQASeniorEngineer, orchestration.RoleValidationSpec,
		orchestration.RoleDocsSpecialist, evidence.Role("gatekeeper"),
	}
	for _, role := range roles {
		dasChdir(t, "demo.md")
		dasInit(t, "demo", role)
		_, e := dasReadJSON(t, "demo", role)
		if len(e.Inputs) != 0 {
			t.Fatalf("%s inputs not empty: %v", role, e.Inputs)
		}
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: PlanInputs is the only source shared with the validator
func TestDAS_PlanInputsSharedWithValidator(t *testing.T) {
	dasChdir(t, "demo.md")
	got := evidence.PlanInputs("demo", orchestration.RoleBigThinker)
	if !reflect.DeepEqual(got, orchestration.RequiredPlanInputs("demo")) {
		t.Fatalf("PlanInputs must equal RequiredPlanInputs verbatim: %v", got)
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: PlanInputs returns nil for a non-plan role
func TestDAS_PlanInputsNilNonPlanRole(t *testing.T) {
	dasChdir(t, "demo.md")
	if got := evidence.PlanInputs("demo", orchestration.RoleSeniorEngineer); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func dasContains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func hasFill(s string) bool { return strings.Contains(s, "<FILL:") }
