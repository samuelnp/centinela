package workflow

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// Back-compat: JSON written before modelRoutes existed loads with no routes, and
// a Save round-trip must not grow the key.
func TestModelRoutes_LegacyJSONRoundTripAddsNoKey(t *testing.T) {
	legacy := `{"feature":"f","currentStep":"plan","steps":{"plan":{"status":"in-progress","completedAt":null}}}`
	var wf Workflow
	if err := json.Unmarshal([]byte(legacy), &wf); err != nil {
		t.Fatalf("legacy JSON must load cleanly: %v", err)
	}
	if wf.ModelRoutes != nil || RouteTiers(&wf) != nil || RouteTiers(nil) != nil {
		t.Fatal("an absent field must mean no routes")
	}
	if _, ok := wf.ModelRouteFor("planner"); ok {
		t.Fatal("no route may be reported for a legacy workflow")
	}
	var nilWF *Workflow
	if _, ok := nilWF.ModelRouteFor("planner"); ok {
		t.Fatal("ModelRouteFor must be nil-safe")
	}
	data, err := json.Marshal(&wf)
	if err != nil || strings.Contains(string(data), "modelRoutes") {
		t.Fatalf("a round trip must not add modelRoutes: %s (%v)", data, err)
	}
}

func TestSetModelRoute_LazyInitAndOverwrite(t *testing.T) {
	wf := New("f")
	wf.SetModelRoute("senior-engineer", ModelRoute{Tier: "balanced", Reason: "trivial", DecidedAt: "2026-07-30T00:00:00Z"})
	wf.SetModelRoute("senior-engineer", ModelRoute{Tier: "reasoning", DecidedAt: "2026-07-30T01:00:00Z"})
	route, ok := wf.ModelRouteFor("senior-engineer")
	if !ok || route.Tier != "reasoning" || route.Reason != "" {
		t.Fatalf("an upgrade must overwrite in place, got %#v", route)
	}
	if tiers := RouteTiers(wf); tiers["senior-engineer"] != "reasoning" {
		t.Fatalf("RouteTiers must project the effective tier, got %#v", tiers)
	}
}

func TestRoleScheduledStep(t *testing.T) {
	wf := New("f")
	if step, ok := RoleScheduledStep(wf, orchestration.RoleSeniorEngineer); !ok || step != "code" {
		t.Fatalf("senior-engineer must be scheduled at code, got %q/%v", step, ok)
	}
	if _, ok := RoleScheduledStep(wf, orchestration.RoleMergeSteward); ok {
		t.Fatal("an out-of-band role must not be scheduled")
	}
	if _, ok := RoleScheduledStep(nil, orchestration.RolePlanner); ok {
		t.Fatal("a nil workflow schedules nothing")
	}
}
