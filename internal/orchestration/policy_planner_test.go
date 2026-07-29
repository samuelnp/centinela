package orchestration

import "testing"

func TestRequiredRoles_PlanIsPlannerOnly(t *testing.T) {
	roles := RequiredRoles("plan")
	if len(roles) != 1 || roles[0] != RolePlanner {
		t.Fatalf("RequiredRoles(plan) = %v, want [planner]", roles)
	}
	for _, r := range roles {
		if r == RoleBigThinker || r == RoleFeatureSpecial {
			t.Fatalf("retired legacy role %q still required at plan", r)
		}
	}
}

func TestRolePlanner_ResolvesToReasoningTier(t *testing.T) {
	if got := DefaultTierForRole(RolePlanner); got != TierReasoning {
		t.Fatalf("planner tier = %q, want %q", got, TierReasoning)
	}
}

// The legacy roles keep their historical tiers so a legacy workflow's model
// routing is unchanged.
func TestLegacyPlanRoles_KeepTiers(t *testing.T) {
	if got := DefaultTierForRole(RoleBigThinker); got != TierReasoning {
		t.Fatalf("big-thinker tier = %q, want reasoning", got)
	}
	if got := DefaultTierForRole(RoleFeatureSpecial); got != TierBalanced {
		t.Fatalf("feature-specialist tier = %q, want balanced", got)
	}
}

func TestAllowedRoleSlugs_IncludesPlannerAndLegacy(t *testing.T) {
	want := map[string]bool{"planner": false, "big-thinker": false, "feature-specialist": false}
	for _, s := range AllowedRoleSlugs() {
		if _, ok := want[s]; ok {
			want[s] = true
		}
	}
	for slug, found := range want {
		if !found {
			t.Errorf("AllowedRoleSlugs missing %q", slug)
		}
	}
}

func TestRequiresPlanSnapshot_CoversPlannerAndLegacy(t *testing.T) {
	for _, r := range []Role{RolePlanner, RoleBigThinker, RoleFeatureSpecial} {
		if !requiresPlanSnapshot(r) {
			t.Errorf("%q should require the plan snapshot", r)
		}
	}
	if requiresPlanSnapshot(RoleSeniorEngineer) {
		t.Error("senior-engineer must not require the plan snapshot")
	}
}

func TestNeedsEdgeCases_PlannerCarriesSpecLens(t *testing.T) {
	if !needsEdgeCases(RolePlanner) {
		t.Fatal("planner carries the spec lens and must require edgeCases")
	}
	if needsEdgeCases(RoleBigThinker) {
		t.Fatal("big-thinker never required edgeCases; legacy behavior must not change")
	}
}

func TestDispatchRoleOutputs_PlannerRequiresPlanOrSpecArtifact(t *testing.T) {
	err := dispatchRoleOutputs("p.json", "f", RolePlanner, []string{"internal/x.go"}, nil)
	if err == nil {
		t.Fatal("planner outputs without a plan/spec artifact must be rejected")
	}
	if err := dispatchRoleOutputs("p.json", "f", RolePlanner, []string{"docs/plans/f.md"}, nil); err != nil {
		t.Fatalf("a real plan artifact must satisfy the planner rule: %v", err)
	}
	if err := dispatchRoleOutputs("p.json", "f", RolePlanner, []string{"specs/f.feature"}, nil); err != nil {
		t.Fatalf("a real spec artifact must satisfy the planner rule: %v", err)
	}
}
