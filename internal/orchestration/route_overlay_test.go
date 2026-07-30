package orchestration

import "testing"

func TestApplyRoutes_ReplacesWholesaleWithoutMutating(t *testing.T) {
	models := RoleModels{
		"senior-engineer": {Overrides: map[string]string{"claude": "pinned-model"}},
		"qa-senior":       {Tier: "reasoning"},
	}
	out := ApplyRoutes(models, map[string]string{"senior-engineer": "balanced"}, nil)
	if out["senior-engineer"].Tier != "balanced" || len(out["senior-engineer"].Overrides) != 0 {
		t.Fatalf("a route to a DIFFERENT tier must replace the entry wholesale, got %#v", out["senior-engineer"])
	}
	if out["qa-senior"].Tier != "reasoning" {
		t.Fatal("un-routed roles must survive untouched")
	}
	if len(models["senior-engineer"].Overrides) != 1 {
		t.Fatal("ApplyRoutes must not mutate the shared input map")
	}
}

func TestApplyRoutes_EmptyAndCorruptRoutes(t *testing.T) {
	models := RoleModels{"qa-senior": {Tier: "reasoning"}}
	if got := ApplyRoutes(models, nil, nil); len(got) != 1 || got["qa-senior"].Tier != "reasoning" {
		t.Fatalf("no routes must be a pass-through, got %#v", got)
	}
	got := ApplyRoutes(models, map[string]string{"qa-senior": "turbo"}, nil)
	if got["qa-senior"].Tier != "reasoning" {
		t.Fatalf("a corrupt tier must be ignored, not applied, got %#v", got["qa-senior"])
	}
	if unknown := ApplyRoutes(models, map[string]string{"wizard": "fast"}, nil); len(unknown) != 1 {
		t.Fatalf("an unknown role slug must be ignored, got %#v", unknown)
	}
	routed := ApplyRoutes(nil, map[string]string{"qa-senior": " Fast "}, nil)
	if routed["qa-senior"].Tier != "fast" {
		t.Fatalf("routes must apply onto an empty model table, got %#v", routed)
	}
}

// E1: the floor must bind where the model is RESOLVED, because the workflow
// state file that carries routes is agent-writable in every step.
func TestApplyRoutes_IgnoresRouteBelowFloor(t *testing.T) {
	models := RoleModels{}
	got := ApplyRoutes(models, map[string]string{"gatekeeper": "fast"}, nil)
	if _, overlaid := got["gatekeeper"]; overlaid {
		t.Fatalf("a hand-written sub-floor route must never be honored, got %#v", got["gatekeeper"])
	}
	if RoleTier(RoleGatekeeper, got) != TierReasoning {
		t.Fatal("the gatekeeper must fall back to its reasoning-tier static resolution")
	}
	// Case does not launder the attack: the key normalizes before the check.
	if mixed := ApplyRoutes(models, map[string]string{"GateKeeper": "fast"}, nil); len(mixed) != 0 {
		t.Fatalf("a mixed-case role key must not bypass the floor, got %#v", mixed)
	}
	// A project that consciously lowers the floor gets the route it asked for.
	lowered := ApplyRoutes(models, map[string]string{"gatekeeper": "fast"}, map[string]string{"gatekeeper": "fast"})
	if lowered["gatekeeper"].Tier != "fast" {
		t.Fatalf("an explicitly lowered floor must admit the route, got %#v", lowered["gatekeeper"])
	}
}

// E4: confirming the tier the directive already printed must not silently
// discard the project's per-runner concrete-model pin.
func TestApplyRoutes_PreservesPerRunnerOverrideOnSameTier(t *testing.T) {
	models := RoleModels{"senior-engineer": {Overrides: map[string]string{"claude": "my-pinned-sonnet-4-9"}}}
	out := ApplyRoutes(models, map[string]string{"senior-engineer": "reasoning"}, nil)
	if out["senior-engineer"].Overrides["claude"] != "my-pinned-sonnet-4-9" {
		t.Fatalf("a same-tier route must preserve the override table, got %#v", out["senior-engineer"])
	}
	if id, ok := ResolveModel(RoleSeniorEngineer, out, nil, RunnerClaude); !ok || id != "my-pinned-sonnet-4-9" {
		t.Fatalf("the pinned model must still resolve, got %q/%v", id, ok)
	}
}

func TestHonoredRouteTier_ParsesClearsFloorOrIsIgnored(t *testing.T) {
	if _, ok := HonoredRouteTier(RoleQASeniorEngineer, "", nil); ok {
		t.Fatal("an empty tier is not a route")
	}
	tier, ok := HonoredRouteTier(RoleGatekeeper, " Reasoning ", nil)
	if !ok || tier != TierReasoning {
		t.Fatalf("a floor-clearing route is honored normalized, got %q/%v", tier, ok)
	}
	if _, ok := HonoredRouteTier(RoleBigThinker, "fast", nil); ok {
		t.Fatal("a retired plan role must inherit the planner floor")
	}
}
