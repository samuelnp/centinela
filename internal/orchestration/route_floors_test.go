package orchestration

import "testing"

func TestTierBelow_TotalOrderWithUnknownTiers(t *testing.T) {
	if !TierBelow(TierFast, TierBalanced) || !TierBelow(TierBalanced, TierReasoning) {
		t.Fatal("tiers must order fast < balanced < reasoning")
	}
	if TierBelow(TierReasoning, TierReasoning) || TierBelow(TierReasoning, TierFast) {
		t.Fatal("below must be strict and directional")
	}
	if !TierBelow(Tier("bogus"), TierFast) {
		t.Fatal("an unknown tier must never rank above a real one")
	}
}

func TestDefaultFloors_ShippedStrictAndCopied(t *testing.T) {
	floors := DefaultFloors()
	if floors[RoleGatekeeper] != TierReasoning || floors[RolePlanner] != TierBalanced {
		t.Fatalf("shipped floors changed: %#v", floors)
	}
	if _, ok := floors[RoleSeniorEngineer]; ok {
		t.Fatal("senior-engineer must be floorless by default")
	}
	floors[RoleGatekeeper] = TierFast
	if DefaultFloors()[RoleGatekeeper] != TierReasoning {
		t.Fatal("DefaultFloors must hand out a copy, not the package map")
	}
}

func TestEffectiveFloor_ConfigReplacesDefault(t *testing.T) {
	tier, ok := EffectiveFloor(RoleGatekeeper, nil)
	if !ok || tier != TierReasoning {
		t.Fatalf("absent config must keep the default floor, got %q/%v", tier, ok)
	}
	tier, ok = EffectiveFloor(RoleGatekeeper, map[string]string{"gatekeeper": "balanced"})
	if !ok || tier != TierBalanced {
		t.Fatalf("an explicit floor must replace the default, got %q/%v", tier, ok)
	}
	tier, ok = EffectiveFloor(RoleGatekeeper, map[string]string{"gatekeeper": "turbo"})
	if !ok || tier != TierReasoning {
		t.Fatalf("an invalid floor must fall back to the default, got %q/%v", tier, ok)
	}
	if _, ok := EffectiveFloor(RoleSeniorEngineer, nil); ok {
		t.Fatal("a floorless role must report no floor")
	}
	if tier, ok := EffectiveFloor(RoleSeniorEngineer, map[string]string{"senior-engineer": "fast"}); !ok || tier != TierFast {
		t.Fatalf("config may introduce a floor for a floorless role, got %q/%v", tier, ok)
	}
}

// E2: a legacy contract schedules the retired roles, which must not run
// floorless just because the shipped table is keyed on their successors.
func TestEffectiveFloor_RetiredRolesInheritTheirSuccessorsFloor(t *testing.T) {
	for _, role := range []Role{RoleBigThinker, RoleFeatureSpecial} {
		if tier, ok := EffectiveFloor(role, nil); !ok || tier != TierBalanced {
			t.Fatalf("%s must inherit the planner floor, got %q/%v", role, tier, ok)
		}
	}
	if tier, ok := EffectiveFloor(RoleValidationSpec, nil); !ok || tier != TierReasoning {
		t.Fatalf("validation-specialist must inherit the gatekeeper floor, got %q/%v", tier, ok)
	}
	// A project lowering the successor's floor lowers the retired role's too.
	tier, ok := EffectiveFloor(RoleBigThinker, map[string]string{"planner": "fast"})
	if !ok || tier != TierFast {
		t.Fatalf("the successor's config floor must reach the retired role, got %q/%v", tier, ok)
	}
	// An explicit entry on the retired role itself still wins.
	tier, ok = EffectiveFloor(RoleBigThinker, map[string]string{"big-thinker": "reasoning", "planner": "fast"})
	if !ok || tier != TierReasoning {
		t.Fatalf("an explicit retired-role floor must win, got %q/%v", tier, ok)
	}
}
