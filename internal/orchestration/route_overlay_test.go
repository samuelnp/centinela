package orchestration

import "testing"

func TestApplyRoutes_ReplacesWholesaleWithoutMutating(t *testing.T) {
	models := RoleModels{
		"senior-engineer": {Overrides: map[string]string{"claude": "pinned-model"}},
		"qa-senior":       {Tier: "reasoning"},
	}
	out := ApplyRoutes(models, map[string]string{"senior-engineer": "balanced"})
	if out["senior-engineer"].Tier != "balanced" || len(out["senior-engineer"].Overrides) != 0 {
		t.Fatalf("a routed tier must replace the role entry wholesale, got %#v", out["senior-engineer"])
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
	if got := ApplyRoutes(models, nil); len(got) != 1 || got["qa-senior"].Tier != "reasoning" {
		t.Fatalf("no routes must be a pass-through, got %#v", got)
	}
	got := ApplyRoutes(models, map[string]string{"qa-senior": "turbo"})
	if got["qa-senior"].Tier != "reasoning" {
		t.Fatalf("a corrupt tier must be ignored, not applied, got %#v", got["qa-senior"])
	}
	routed := ApplyRoutes(nil, map[string]string{"planner": " Fast "})
	if routed["planner"].Tier != "fast" {
		t.Fatalf("routes must apply onto an empty model table, got %#v", routed)
	}
}
