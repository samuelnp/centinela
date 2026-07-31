package workflow

import "testing"

func collidedWorkflow() *Workflow {
	return &Workflow{Feature: "f", ModelRoutes: map[string]ModelRoute{
		"Senior-Engineer": {Tier: "fast", Reason: "hand-written"},
		"senior-engineer": {Tier: "balanced", Reason: "hand-written"},
		"qa-senior":       {Tier: "fast", Reason: "cheap tier"},
	}}
}

// `route show` reads routes through the same canonical view the overlay uses, so
// a collided role must render as having no route rather than claiming one the
// hook will not honor.
func TestModelRouteFor_AmbiguousCollisionHasNoRoute(t *testing.T) {
	wf := collidedWorkflow()
	if _, ok := wf.ModelRouteFor("senior-engineer"); ok {
		t.Fatal("a collided role must report no effective route")
	}
	route, ok := wf.ModelRouteFor("qa-senior")
	if !ok || route.Tier != "fast" {
		t.Fatalf("an unambiguous route must still resolve, got %+v ok=%v", route, ok)
	}
}

func TestRouteTiers_DropsCollisionAndNormalizesKeys(t *testing.T) {
	got := RouteTiers(collidedWorkflow())
	if _, ok := got["senior-engineer"]; ok {
		t.Fatalf("collided role must not reach the overlay, got %v", got)
	}
	if got["qa-senior"] != "fast" {
		t.Fatalf("expected the surviving route, got %v", got)
	}
}

// A route recorded under a non-canonical key alone (no collision) still applies:
// dropping it would silently ignore a decision the operator did record.
func TestRouteTiers_NormalizesLoneNonCanonicalKey(t *testing.T) {
	wf := &Workflow{Feature: "f", ModelRoutes: map[string]ModelRoute{
		"QA-Senior": {Tier: "balanced"},
	}}
	if got := RouteTiers(wf); got["qa-senior"] != "balanced" {
		t.Fatalf("expected the normalized key to carry the tier, got %v", got)
	}
}
