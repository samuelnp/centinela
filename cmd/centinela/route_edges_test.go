package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestPrintRoutingDirective_DynamicOnlyAndSilentWhenRouted(t *testing.T) {
	wf := routeRepo(t, "plan", dynamicToml)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	out := captureStdout(t, func() { printRoutingDirective(wf, cfg) })
	if !strings.Contains(out, "routing (dynamic): unrouted [planner]") {
		t.Fatalf("start must hand the decision to the orchestrator: %s", out)
	}
	wf.SetModelRoute("planner", workflow.ModelRoute{Tier: "reasoning", DecidedAt: "2026-07-30T00:00:00Z"})
	if out := captureStdout(t, func() { printRoutingDirective(wf, cfg) }); out != "" {
		t.Fatalf("a routed first step must print nothing, got %q", out)
	}
	if out := captureStdout(t, func() { printRoutingDirective(wf, &config.Config{}) }); out != "" {
		t.Fatalf("static mode must print nothing, got %q", out)
	}
}

// A corrupt recorded tier must not read as an upgrade: the comparison falls back
// to the static default so rule 6 still refuses the downgrade.
func TestCurrentEffectiveTier_CorruptRouteFallsBackToStatic(t *testing.T) {
	wf := routeRepo(t, "code", dynamicToml)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	wf.SetModelRoute("senior-engineer", workflow.ModelRoute{Tier: "turbo"})
	req := buildRouteRequest(wf, cfg, orchestration.RoleSeniorEngineer, orchestration.TierFast, "x")
	if req.CurrentTier != orchestration.TierReasoning {
		t.Fatalf("a corrupt tier must fall back to the static default, got %q", req.CurrentTier)
	}
}

func TestLoadRoutingConfig_SurfacesConfigErrors(t *testing.T) {
	routeRepo(t, "plan", "[orchestration]\nrouting_mode = \"turbo\"\n")
	if _, err := loadRoutingConfig(); err == nil || !strings.Contains(err.Error(), "routing_mode") {
		t.Fatalf("a bad routing_mode must fail at load, got %v", err)
	}
}

func TestRunRouteShow_MissingWorkflowAndUnwritableState(t *testing.T) {
	routeRepo(t, "plan", dynamicToml)
	if err := runRouteShow(nil, []string{"ghost"}); err == nil {
		t.Fatal("an unknown feature must surface the missing-workflow error")
	}
	routeSetReason = "config-only change"
	os.Chmod(workflow.FilePath("f"), 0o400)                       //nolint:errcheck
	t.Cleanup(func() { os.Chmod(workflow.FilePath("f"), 0o644) }) //nolint:errcheck
	if err := runRouteSet(nil, []string{"f", "senior-engineer", "balanced"}); err == nil ||
		!strings.Contains(err.Error(), "cannot save workflow") {
		t.Fatalf("a failed save must surface, got %v", err)
	}
}
