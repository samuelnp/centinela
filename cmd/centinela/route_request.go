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
	floors := config.OrchestrationFloors(cfg)
	floor, hasFloor := orchestration.EffectiveFloor(role, floors)
	step, scheduled := workflow.RoleScheduledStep(wf, role)
	return orchestration.RouteRequest{
		Role:         role,
		NewTier:      tier,
		CurrentTier:  currentEffectiveTier(wf, role, static, floors),
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
// route wins ONLY where the overlay would honor it, else the static default. A
// corrupt or below-floor recorded tier therefore cannot fake an upgrade past
// rule 6 — the comparison matches what the hook actually emits.
func currentEffectiveTier(wf *workflow.Workflow, role orchestration.Role, static orchestration.Tier, floors map[string]string) orchestration.Tier {
	route, ok := wf.ModelRouteFor(string(role))
	if !ok {
		return static
	}
	if tier, honored := orchestration.HonoredRouteTier(role, route.Tier, floors); honored {
		return tier
	}
	return static
}
