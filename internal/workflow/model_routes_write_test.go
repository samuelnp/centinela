package workflow

import "testing"

// Regression: the write must clear a pre-existing key naming the same role under
// different casing. Leaving it made resolution drop BOTH entries, so `route set`
// reported success and changed nothing, and re-running never converged.
func TestSetModelRoute_ReplacesNonCanonicalKeyForSameRole(t *testing.T) {
	wf := &Workflow{Feature: "f", ModelRoutes: map[string]ModelRoute{
		"GateKeeper": {Tier: "balanced", Reason: "hand-written"},
	}}
	wf.SetModelRoute("gatekeeper", ModelRoute{Tier: "reasoning", Reason: "restored"})

	if len(wf.ModelRoutes) != 1 {
		t.Fatalf("exactly one key per role must survive the write, got %v", wf.ModelRoutes)
	}
	route, ok := wf.ModelRouteFor("gatekeeper")
	if !ok {
		t.Fatal("the route just written must take effect")
	}
	if route.Tier != "reasoning" || route.Reason != "restored" {
		t.Fatalf("the new decision must win, got %+v", route)
	}
	if RouteTiers(wf)["gatekeeper"] != "reasoning" {
		t.Fatalf("the overlay must see the written tier, got %v", RouteTiers(wf))
	}
}

// An ordinary overwrite still replaces in place.
func TestSetModelRoute_OverwritesCanonicalKeyInPlace(t *testing.T) {
	wf := &Workflow{Feature: "f"}
	wf.SetModelRoute("qa-senior", ModelRoute{Tier: "fast"})
	wf.SetModelRoute("qa-senior", ModelRoute{Tier: "balanced"})
	if len(wf.ModelRoutes) != 1 || wf.ModelRoutes["qa-senior"].Tier != "balanced" {
		t.Fatalf("expected one upgraded entry, got %v", wf.ModelRoutes)
	}
}

func TestRawModelRouteRecorded_SeesDroppedAndCanonicalKeys(t *testing.T) {
	wf := &Workflow{Feature: "f", ModelRoutes: map[string]ModelRoute{
		"Senior-Engineer": {Tier: "fast"},
		"senior-engineer": {Tier: "balanced"},
		"qa-senior":       {Tier: "fast"},
	}}
	if _, ok := wf.ModelRouteFor("senior-engineer"); ok {
		t.Fatal("the collided role must still have no effective route")
	}
	if !wf.RawModelRouteRecorded("senior-engineer") {
		t.Fatal("a dropped route must remain visible to the table as recorded")
	}
	if !wf.RawModelRouteRecorded("qa-senior") {
		t.Fatal("an honored route is recorded too")
	}
	if wf.RawModelRouteRecorded("gatekeeper") {
		t.Fatal("a role with no route must not report one")
	}
}
