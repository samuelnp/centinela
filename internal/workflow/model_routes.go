package workflow

import (
	"os"

	"github.com/samuelnp/centinela/internal/orchestration"
)

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

// RoleStepUnderway reports whether delegation for the role's scheduled step has
// begun. A pure "in-progress" status cannot answer this: the first step is
// in-progress from `start`, which is exactly the sanctioned routing window.
// Underway = the step is behind the cursor (or the workflow is done), or it is
// the current step AND an evidence artifact exists — `evidence init` writes both
// stubs, so either one proves delegation began.
func RoleStepUnderway(wf *Workflow, role orchestration.Role) bool {
	step, ok := RoleScheduledStep(wf, role)
	if !ok {
		return false
	}
	if wf.CurrentStep == "done" {
		return true
	}
	order := wf.OrderedSteps()
	if stepIndexIn(step, order) < stepIndexIn(wf.CurrentStep, order) {
		return true
	}
	if step != wf.CurrentStep {
		return false
	}
	return roleEvidenceExists(wf.Feature, role)
}

func roleEvidenceExists(feature string, role orchestration.Role) bool {
	for _, path := range []string{orchestration.MarkdownPath(feature, role), orchestration.JSONPath(feature, role)} {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}
