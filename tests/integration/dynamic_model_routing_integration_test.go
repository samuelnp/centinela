package integration_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// dmrProject chdirs into a temp project with .workflow/ and a centinela.toml.
func dmrProject(t *testing.T, toml string) string {
	t.Helper()
	dir := t.TempDir()
	origin, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origin) })              //nolint:errcheck
	os.Chdir(dir)                                       //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0o755)            //nolint:errcheck
	os.WriteFile("centinela.toml", []byte(toml), 0o644) //nolint:errcheck
	return dir
}

// The full lifecycle against real state on disk: record → persist → reload →
// project → overlay, with the config floors travelling all the way through.
func TestRouting_LifecycleSetPersistOverlayRoundTrip(t *testing.T) {
	dmrProject(t, "[orchestration]\nrouting_mode = \"dynamic\"\n")
	wf := workflow.New("f")
	wf.CurrentStep = "code"
	wf.SetModelRoute("senior-engineer", workflow.ModelRoute{Tier: "balanced", Reason: "config-only", DecidedAt: "2026-01-02T00:00:00Z"})
	if err := workflow.Save(wf); err != nil {
		t.Fatalf("save: %v", err)
	}
	reloaded, err := workflow.Load("f")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	route, ok := reloaded.ModelRouteFor("senior-engineer")
	if !ok || route.Tier != "balanced" || route.Reason != "config-only" {
		t.Fatalf("the route must survive the round-trip, got %#v/%v", route, ok)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	if !config.DynamicRoutingEnabled(cfg) {
		t.Fatal("routing_mode = dynamic must enable the overlay")
	}
	out := orchestration.ApplyRoutes(nil, workflow.RouteTiers(reloaded), config.OrchestrationFloors(cfg))
	if id, ok := orchestration.ResolveModel(orchestration.RoleSeniorEngineer, out, nil, orchestration.RunnerClaude); !ok || id != "sonnet" {
		t.Fatalf("the routed tier must reach model resolution, got %q/%v", id, ok)
	}
}

// E1 end-to-end at the package boundary: hand-written state, real Load, real
// overlay — the gatekeeper's resolved model must not move.
func TestRouting_HandWrittenRouteCannotWeakenGatekeeper(t *testing.T) {
	dmrProject(t, "[orchestration]\nrouting_mode = \"dynamic\"\n")
	body := `{"feature":"g","currentStep":"validate","steps":{},` +
		`"modelRoutes":{"gatekeeper":{"tier":"fast","decidedAt":"2026-01-02T00:00:00Z"}}}`
	if err := os.WriteFile(workflow.FilePath("g"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	wf, err := workflow.Load("g")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	cfg, _ := config.Load()
	out := orchestration.ApplyRoutes(nil, workflow.RouteTiers(wf), config.OrchestrationFloors(cfg))
	id, ok := orchestration.ResolveModel(orchestration.RoleGatekeeper, out, nil, orchestration.RunnerClaude)
	if !ok || id != "opus" {
		t.Fatalf("the verifier must stay reasoning-tier, got %q/%v", id, ok)
	}
}

// E3 — a floor above the static default is legal (the project's explicit
// [orchestration.models] choice is sanctioned), so the fix is that the SURFACE
// says what the floor governs instead of claiming an enforcement it never does.
func TestRouting_FloorAboveStaticIsLabelledRoutesOnly(t *testing.T) {
	dmrProject(t, "[orchestration]\nrouting_mode = \"dynamic\"\n\n"+
		"[orchestration.models]\nqa-senior = \"balanced\"\n\n"+
		"[orchestration.floors]\nqa-senior = \"reasoning\"\n")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	floors := config.OrchestrationFloors(cfg)
	models := orchestration.RoleModels{"qa-senior": {Tier: "balanced"}}
	if tier := orchestration.RoleTier(orchestration.RoleQASeniorEngineer, models); tier != orchestration.TierBalanced {
		t.Fatalf("a floor must not silently raise the static tier, got %q", tier)
	}
	line := orchestration.RoutingDirective("f", []orchestration.Role{orchestration.RoleQASeniorEngineer}, nil, floors, models)
	if !strings.Contains(line, "floors (routes only): qa-senior>=reasoning") {
		t.Fatalf("the directive must scope its floor claim to routes: %s", line)
	}
	// And the floor really does bind the route path, in both directions.
	if out := orchestration.ApplyRoutes(models, map[string]string{"qa-senior": "balanced"}, floors); out["qa-senior"].Tier != "balanced" {
		t.Fatalf("a sub-floor route must leave the static entry in place, got %#v", out["qa-senior"])
	}
	if out := orchestration.ApplyRoutes(models, map[string]string{"qa-senior": "reasoning"}, floors); out["qa-senior"].Tier != "reasoning" {
		t.Fatalf("a floor-clearing route must apply, got %#v", out["qa-senior"])
	}
}

// Back-compat, both directions of the schema boundary.
func TestRouting_LegacyJSONWithoutModelRoutesRoundTrips(t *testing.T) {
	dmrProject(t, "")
	body := `{"feature":"old","currentStep":"code","steps":{"plan":{"status":"completed","completedAt":null}}}`
	if err := os.WriteFile(workflow.FilePath("old"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	wf, err := workflow.Load("old")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if wf.ModelRoutes != nil || workflow.RouteTiers(wf) != nil {
		t.Fatalf("absent modelRoutes must load as nil, got %#v", wf.ModelRoutes)
	}
	if err := workflow.Save(wf); err != nil {
		t.Fatalf("save: %v", err)
	}
	data, _ := os.ReadFile(workflow.FilePath("old"))
	if strings.Contains(string(data), "modelRoutes") {
		t.Fatalf("omitempty must keep the key absent after a round-trip: %s", data)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("the re-saved state must stay valid JSON: %v", err)
	}
}

// E11 — flipping dynamic → static → dynamic must not lose or silently mutate
// recorded routes, and static mode must ignore them outright.
func TestRouting_DynamicStaticDynamicRoundTrip(t *testing.T) {
	dmrProject(t, "[orchestration]\nrouting_mode = \"dynamic\"\n")
	wf := workflow.New("f")
	wf.SetModelRoute("qa-senior", workflow.ModelRoute{Tier: "fast", Reason: "trivial", DecidedAt: "2026-01-02T00:00:00Z"})
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("centinela.toml", []byte("[orchestration]\nrouting_mode = \"static\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, _ := config.Load()
	if config.DynamicRoutingEnabled(cfg) {
		t.Fatal("static mode must disable the overlay")
	}
	reloaded, err := workflow.Load("f")
	if err != nil {
		t.Fatal(err)
	}
	if route, ok := reloaded.ModelRouteFor("qa-senior"); !ok || route.Tier != "fast" {
		t.Fatalf("static mode must preserve recorded routes untouched, got %#v/%v", route, ok)
	}
	if err := os.WriteFile("centinela.toml", []byte("[orchestration]\nrouting_mode = \"dynamic\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, _ = config.Load()
	out := orchestration.ApplyRoutes(nil, workflow.RouteTiers(reloaded), config.OrchestrationFloors(cfg))
	if out["qa-senior"].Tier != "fast" {
		t.Fatalf("flipping back must re-arm the same route, got %#v", out["qa-senior"])
	}
}
