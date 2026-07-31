package workflow

import "github.com/samuelnp/centinela/internal/orchestration"

// SetModelRoute records (or replaces) the route for a role, initializing the map
// lazily so a workflow loaded from legacy JSON needs no migration.
//
// The write NORMALIZES, mirroring the read. A state file already naming this
// role under a different casing would otherwise collide with the new entry, and
// resolution drops both — so `route set` would report success, emit its
// telemetry, and change nothing, no matter how many times it was re-run. Leaving
// exactly one key per role is what makes the command's success truthful.
func (wf *Workflow) SetModelRoute(role string, route ModelRoute) {
	if wf.ModelRoutes == nil {
		wf.ModelRoutes = map[string]ModelRoute{}
	}
	key := role
	if normalized, ok := orchestration.NormalizeRole(role); ok {
		key = string(normalized)
		for raw := range wf.ModelRoutes {
			if other, known := orchestration.NormalizeRole(raw); known && string(other) == key && raw != key {
				delete(wf.ModelRoutes, raw)
			}
		}
	}
	wf.ModelRoutes[key] = route
}

// RawModelRouteRecorded reports whether the state names this role at all,
// including under a key that resolution drops. `route show` needs it to render a
// dropped route as "ignored" rather than as "static", which would erase every
// trace that a decision was ever recorded — the only surface able to explain why
// the emitted model does not match it.
func (wf *Workflow) RawModelRouteRecorded(role string) bool {
	if wf == nil {
		return false
	}
	for raw := range wf.ModelRoutes {
		if other, known := orchestration.NormalizeRole(raw); known && string(other) == role {
			return true
		}
	}
	return false
}
