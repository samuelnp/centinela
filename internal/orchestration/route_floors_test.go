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
