package evidence

import (
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestAllRolesCoversContractRoster(t *testing.T) {
	want := []Role{
		orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial,
		orchestration.RoleSeniorEngineer, orchestration.RoleUXUISpecialist,
		orchestration.RoleQASeniorEngineer, orchestration.RoleValidationSpec,
		orchestration.RoleDocsSpecialist, orchestration.RoleMergeSteward,
		"gatekeeper", "production-readiness",
	}
	got := AllRoles()
	if len(got) != len(want) {
		t.Fatalf("count: got %d want %d", len(got), len(want))
	}
	for i, r := range want {
		if got[i] != r {
			t.Errorf("[%d] got %s want %s", i, got[i], r)
		}
	}
}

func TestStepForRoleCoverage(t *testing.T) {
	cases := map[Role]string{
		orchestration.RoleBigThinker:       "plan",
		orchestration.RoleSeniorEngineer:   "code",
		orchestration.RoleQASeniorEngineer: "tests",
		orchestration.RoleValidationSpec:   "validate",
		orchestration.RoleDocsSpecialist:   "docs",
		orchestration.RoleMergeSteward:     "merge",
		Role("gatekeeper"):                 "validate",
		Role("production-readiness"):       "validate",
		Role("ghost"):                      "",
	}
	for r, want := range cases {
		if got := stepForRole(r); got != want {
			t.Errorf("stepForRole(%q) = %q, want %q", r, got, want)
		}
	}
}

func TestHandoffForRoleCoverage(t *testing.T) {
	cases := map[Role]string{
		orchestration.RoleBigThinker:       "feature-specialist",
		orchestration.RoleFeatureSpecial:   "senior-engineer",
		orchestration.RoleSeniorEngineer:   "qa-senior",
		orchestration.RoleUXUISpecialist:   "qa-senior",
		orchestration.RoleQASeniorEngineer: "validation-specialist",
		orchestration.RoleValidationSpec:   "documentation-specialist",
		orchestration.RoleDocsSpecialist:   "complete",
		orchestration.RoleMergeSteward:     "complete",
		Role("ghost"):                      "complete",
	}
	for r, want := range cases {
		if got := handoffForRole(r); got != want {
			t.Errorf("handoffForRole(%q) = %q, want %q", r, got, want)
		}
	}
}
