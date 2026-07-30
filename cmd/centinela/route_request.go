package main

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// buildRouteRequest resolves the facts orchestration.ValidateRoute judges. Pure
// data gathering — every rule and every default is owned by the domain (G7).
func buildRouteRequest(wf *workflow.Workflow, cfg *config.Config, role orchestration.Role, tier orchestration.Tier, reason string) orchestration.RouteRequest {
	models, _ := orchestrationRouting(cfg)
	static := orchestration.RoleTier(role, models)
	floor, hasFloor := orchestration.EffectiveFloor(role, config.OrchestrationFloors(cfg))
	step, scheduled := workflow.RoleScheduledStep(wf, role)
	return orchestration.RouteRequest{
		Role:         role,
		NewTier:      tier,
		CurrentTier:  currentEffectiveTier(wf, role, static),
		StaticTier:   static,
		Floor:        floor,
		HasFloor:     hasFloor,
		Reason:       reason,
		Step:         step,
		Scheduled:    scheduled,
		StepUnderway: workflow.RoleStepUnderway(wf, role),
	}
}

// currentEffectiveTier is the tier a role resolves to TODAY: an already-recorded
// route wins, else the static default. A corrupt recorded tier falls back to the
// static default so a bad state file cannot fake an upgrade past rule 6.
func currentEffectiveTier(wf *workflow.Workflow, role orchestration.Role, static orchestration.Tier) orchestration.Tier {
	route, ok := wf.ModelRouteFor(string(role))
	if !ok {
		return static
	}
	if tier, valid := orchestration.NormalizeTier(route.Tier); valid {
		return tier
	}
	return static
}
