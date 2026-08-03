package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// roleNames flattens a role slice for comparison.
func roleNames(roles []orchestration.Role) string {
	out := ""
	for i, r := range roles {
		if i > 0 {
			out += ","
		}
		out += string(r)
	}
	return out
}

// alternateContractRoles is the ONLY widening the handoff gate allows, so what
// it returns per step is worth pinning at the source: exactly the other arm of
// each contract-pin branch, and NOTHING for every step that has only one legal
// occupant. A step outside the two pinned ones must return nil, or the
// tolerance would silently start admitting roles from steps it never covered.
func TestAlternateContractRolesMirrorsOnlyThePinnedBranches(t *testing.T) {
	for _, tc := range []struct {
		name   string
		legacy bool
		step   string
		want   string
	}{
		{"validate under adversarial pin", false, "validate", "validation-specialist"},
		{"validate under legacy pin", true, "validate", "gatekeeper"},
		{"plan under planner pin", false, "plan", "big-thinker,feature-specialist"},
		{"plan under legacy pin", true, "plan", "planner"},
		{"code has one occupant", false, "code", ""},
		{"tests has one occupant", false, "tests", ""},
		{"docs has one occupant", false, "docs", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hoRepo(t, "f", hoCase{legacyValidate: tc.legacy, legacyPlan: tc.legacy})
			if got := roleNames(alternateContractRoles("f", tc.step)); got != tc.want {
				t.Fatalf("alternateContractRoles(%s) = %q, want %q", tc.step, got, tc.want)
			}
		})
	}
}

// An empty handoffTo belongs to ValidateEvidence's incomplete-fields rule; the
// chain check must stay silent on it rather than derive a successor for a
// field nobody set.
func TestCheckHandoffToIgnoresAnUnsetField(t *testing.T) {
	hoRepo(t, "f", hoCase{})
	if err := CheckHandoffTo("f", "tests", orchestration.RoleQASeniorEngineer, ""); err != nil {
		t.Fatalf("an unset handoffTo is not this rule's business, got %v", err)
	}
}
