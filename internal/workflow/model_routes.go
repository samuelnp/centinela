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

// SetModelRoute and RawModelRouteRecorded live in model_routes_write.go.

// canonicalRoutes folds the raw state map onto normalized role keys, applying
// orchestration.CanonicalRouteTiers as the single authority on which keys
// survive. Reading every route through it is what keeps the `route show` table
// and the model the hook actually emits from ever disagreeing.
func canonicalRoutes(wf *Workflow) map[string]ModelRoute {
	if wf == nil || len(wf.ModelRoutes) == 0 {
		return nil
	}
	tiers := make(map[string]string, len(wf.ModelRoutes))
	for raw, route := range wf.ModelRoutes {
		tiers[raw] = route.Tier
	}
	kept := orchestration.CanonicalRouteTiers(tiers)
	out := make(map[string]ModelRoute, len(kept))
	for raw, route := range wf.ModelRoutes {
		role, known := orchestration.NormalizeRole(raw)
		if !known {
			continue
		}
		if _, survives := kept[string(role)]; survives {
			out[string(role)] = route
		}
	}
	return out
}

// ModelRouteFor returns the role's recorded route, if any. Nil-safe. A role
// whose key collides ambiguously in the raw state has no effective route.
func (wf *Workflow) ModelRouteFor(role string) (ModelRoute, bool) {
	if wf == nil {
		return ModelRoute{}, false
	}
	route, ok := canonicalRoutes(wf)[role]
	return route, ok
}

// RouteTiers projects the recorded routes onto the role→tier shape
// orchestration.ApplyRoutes consumes. Nil for a workflow with no routes.
func RouteTiers(wf *Workflow) map[string]string {
	canonical := canonicalRoutes(wf)
	if len(canonical) == 0 {
		return nil
	}
	out := make(map[string]string, len(canonical))
	for role, route := range canonical {
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
