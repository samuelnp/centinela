package unit_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// E1 — the single highest-value assertion in this suite. `.workflow/<f>.json` is
// agent-writable in every step, so a floor that lives only in `route set` is not
// a floor at all: the overlay that RESOLVES the model must refuse it too.
func TestApplyRoutes_RefusesRouteBelowFloor(t *testing.T) {
	handWritten := map[string]string{"gatekeeper": "fast"}
	out := orchestration.ApplyRoutes(orchestration.RoleModels{}, handWritten, nil)
	if orchestration.RoleTier(orchestration.RoleGatekeeper, out) != orchestration.TierReasoning {
		t.Fatalf("a sub-floor route must fall back to the static tier, got %#v", out)
	}
	id, ok := orchestration.ResolveModel(orchestration.RoleGatekeeper, out, nil, orchestration.RunnerClaude)
	if !ok || id != "opus" {
		t.Fatalf("the verifier's model must not be downgraded, got %q/%v", id, ok)
	}
	// Same posture as an unparseable tier: ignored, never fatal.
	if corrupt := orchestration.ApplyRoutes(nil, map[string]string{"gatekeeper": "ultra"}, nil); len(corrupt) != 0 {
		t.Fatalf("a corrupt tier must be ignored, got %#v", corrupt)
	}
	// A route that CLEARS the floor is still honored — the fix bounds, not bans.
	honored := orchestration.ApplyRoutes(nil, map[string]string{"qa-senior": "fast"}, nil)
	if honored["qa-senior"].Tier != "fast" {
		t.Fatalf("a floorless role must still be routable, got %#v", honored)
	}
}

// E1 (config-lowered floor) — a floor a project explicitly lowers in
// [orchestration.floors] is the floor the overlay enforces, no more, no less.
func TestApplyRoutes_HonorsConfigLoweredFloor(t *testing.T) {
	floors := map[string]string{"gatekeeper": "fast", "qa-senior": "reasoning"}
	out := orchestration.ApplyRoutes(nil, map[string]string{
		"gatekeeper": "fast", "qa-senior": "balanced",
	}, floors)
	if out["gatekeeper"].Tier != "fast" {
		t.Fatalf("a consciously lowered floor must admit the route, got %#v", out["gatekeeper"])
	}
	if _, routed := out["qa-senior"]; routed {
		t.Fatalf("a raised floor must reject a route beneath it, got %#v", out["qa-senior"])
	}
}

// E4 — `route set <f> senior-engineer reasoning` is a visible no-op in
// `route show`; it must also be a no-op for the project's per-runner pin.
func TestApplyRoutes_PreservesPerRunnerOverride(t *testing.T) {
	models := orchestration.RoleModels{
		"senior-engineer": {Overrides: map[string]string{"claude": "my-pinned-sonnet-4-9"}},
	}
	same := orchestration.ApplyRoutes(models, map[string]string{"senior-engineer": "reasoning"}, nil)
	id, ok := orchestration.ResolveModel(orchestration.RoleSeniorEngineer, same, nil, orchestration.RunnerClaude)
	if !ok || id != "my-pinned-sonnet-4-9" {
		t.Fatalf("a same-tier route must keep the pin, got %q/%v", id, ok)
	}
	// The other direction: a real tier change legitimately replaces the entry.
	moved := orchestration.ApplyRoutes(models, map[string]string{"senior-engineer": "balanced"}, nil)
	id, ok = orchestration.ResolveModel(orchestration.RoleSeniorEngineer, moved, nil, orchestration.RunnerClaude)
	if !ok || id != "sonnet" {
		t.Fatalf("a different-tier route must replace the entry, got %q/%v", id, ok)
	}
	if len(models["senior-engineer"].Overrides) != 1 {
		t.Fatal("ApplyRoutes must never mutate the shared static table")
	}
}

// E2 — a legacy plan contract schedules the retired roles, which must inherit
// the planner floor rather than run floorless.
func TestEffectiveFloor_LegacyPlanRolesInheritPlannerFloor(t *testing.T) {
	for _, role := range []orchestration.Role{orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial} {
		tier, ok := orchestration.EffectiveFloor(role, nil)
		if !ok || tier != orchestration.TierBalanced {
			t.Fatalf("%s must inherit planner>=balanced, got %q/%v", role, tier, ok)
		}
		if out := orchestration.ApplyRoutes(nil, map[string]string{string(role): "fast"}, nil); len(out) != 0 {
			t.Fatalf("%s must not be routable below the planner floor, got %#v", role, out)
		}
	}
}

// E2 (validate half) — a legacy validate contract schedules
// validation-specialist and no gatekeeper row at all; without the successor
// alias the whole validate step is floorless.
func TestEffectiveFloor_LegacyValidateRoleInheritsGatekeeperFloor(t *testing.T) {
	tier, ok := orchestration.EffectiveFloor(orchestration.RoleValidationSpec, nil)
	if !ok || tier != orchestration.TierReasoning {
		t.Fatalf("validation-specialist must inherit gatekeeper>=reasoning, got %q/%v", tier, ok)
	}
	lowered := map[string]string{"gatekeeper": "fast"}
	if tier, _ := orchestration.EffectiveFloor(orchestration.RoleValidationSpec, lowered); tier != orchestration.TierFast {
		t.Fatalf("lowering the successor's floor must reach the retired role, got %q", tier)
	}
}
