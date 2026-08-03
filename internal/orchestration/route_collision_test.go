package orchestration

import "testing"

// A raw state map can name one role twice with different casing. Resolving it in
// map order made the emitted model differ between identical runs; the collision
// must be dropped instead.
func TestCanonicalRouteTiers_DropsAmbiguousCollision(t *testing.T) {
	got := CanonicalRouteTiers(map[string]string{
		"Senior-Engineer": "fast",
		"senior-engineer": "reasoning",
		"qa-senior":       "balanced",
		"not-a-role":      "fast",
	})
	if _, ok := got["senior-engineer"]; ok {
		t.Fatal("an ambiguous casing collision must resolve to no route at all")
	}
	if got["qa-senior"] != "balanced" {
		t.Fatalf("an unambiguous route must survive, got %q", got["qa-senior"])
	}
	if len(got) != 1 {
		t.Fatalf("unknown roles must be dropped too, got %v", got)
	}
}

// The overlay must be a pure function of the state: same input, same output,
// every run. Repeating it exercises Go's randomized map iteration order.
func TestApplyRoutes_DeterministicUnderDuplicateCasing(t *testing.T) {
	models := RoleModels{"senior-engineer": RoleModel{Tier: "reasoning"}}
	routes := map[string]string{
		"Senior-Engineer": "fast",
		"senior-engineer": "balanced",
	}
	first := ApplyRoutes(models, routes, nil)["senior-engineer"].Tier
	for i := 0; i < 64; i++ {
		if got := ApplyRoutes(models, routes, nil)["senior-engineer"].Tier; got != first {
			t.Fatalf("run %d resolved %q, first run resolved %q", i, got, first)
		}
	}
	if first != "reasoning" {
		t.Fatalf("a collided role must fall back to its static tier, got %q", first)
	}
}

// The floor still binds when the colliding keys agree on a sub-floor tier.
func TestApplyRoutes_CollisionCannotLaunderASubFloorTier(t *testing.T) {
	models := RoleModels{"gatekeeper": RoleModel{Tier: "reasoning"}}
	floors := map[string]string{"gatekeeper": "reasoning"}
	routes := map[string]string{"GateKeeper": "fast", "gatekeeper": "fast"}
	if got := ApplyRoutes(models, routes, floors)["gatekeeper"]; got.Tier != "reasoning" {
		t.Fatalf("the verifier must stay at its floor, got %q", got.Tier)
	}
}
