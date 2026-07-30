package workflow

import "github.com/samuelnp/centinela/internal/orchestration"

// ModelRoute is one role's recorded routing decision. Exactly one route is
// effective per role: an upgrade overwrites in place, and the append-only
// history lives in telemetry rather than in this state file.
type ModelRoute struct {
	Tier      string `json:"tier"`
	Reason    string `json:"reason,omitempty"`
	DecidedAt string `json:"decidedAt"` // RFC3339 UTC, like StepState.CompletedAt
}

// SetModelRoute records (or replaces) the route for a role, initializing the
// map lazily so a workflow loaded from legacy JSON needs no migration.
func (wf *Workflow) SetModelRoute(role string, route ModelRoute) {
	if wf.ModelRoutes == nil {
		wf.ModelRoutes = map[string]ModelRoute{}
	}
	wf.ModelRoutes[role] = route
}

// ModelRouteFor returns the role's recorded route, if any. Nil-safe.
func (wf *Workflow) ModelRouteFor(role string) (ModelRoute, bool) {
	if wf == nil {
		return ModelRoute{}, false
	}
	route, ok := wf.ModelRoutes[role]
	return route, ok
}

// RouteTiers projects the recorded routes onto the role→tier shape
// orchestration.ApplyRoutes consumes. Nil for a workflow with no routes.
func RouteTiers(wf *Workflow) map[string]string {
	if wf == nil || len(wf.ModelRoutes) == 0 {
		return nil
	}
	out := make(map[string]string, len(wf.ModelRoutes))
	for role, route := range wf.ModelRoutes {
		out[role] = route.Tier
	}
	return out
}

// RoleScheduledStep finds the step whose evidence this workflow requires from
// the role. It scans the workflow's OWN step order through the contract-aware
// RequiredEvidenceRoles, so archetype subsets, reorders, non-user-facing skips
// and legacy contracts are all honored without a parallel role↔step table.
func RoleScheduledStep(wf *Workflow, role orchestration.Role) (string, bool) {
	if wf == nil {
		return "", false
	}
	for _, step := range wf.OrderedSteps() {
		for _, required := range RequiredEvidenceRoles(wf.Feature, step) {
			if required == role {
				return step, true
			}
		}
	}
	return "", false
}

// RoleStepUnderway lives in model_routes_underway.go.
