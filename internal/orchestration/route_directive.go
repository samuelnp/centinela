package orchestration

import (
	"fmt"
	"strings"
)

// RoutingDirective renders the ONE dynamic-mode line handing the routing
// decision to the orchestrator: which of this step's roles are still un-routed,
// their floors (only where one applies), the static defaults as reference, and
// the command that records the decision.
//
// It returns "" once every role is routed — the routed tiers are already visible
// in the directive's existing model: annotations, so repeating them would be
// pure noise. models must be the PRE-overlay (static) table so the reference
// column keeps showing what the decision is being measured against.
func RoutingDirective(feature string, roles []Role, routed, floors map[string]string, models RoleModels) string {
	unrouted := unroutedRoles(roles, routed)
	if len(unrouted) == 0 {
		return ""
	}
	names := make([]string, 0, len(unrouted))
	floorParts := make([]string, 0, len(unrouted))
	staticParts := make([]string, 0, len(unrouted))
	for _, role := range unrouted {
		names = append(names, string(role))
		if floor, ok := EffectiveFloor(role, floors); ok {
			floorParts = append(floorParts, fmt.Sprintf("%s>=%s", role, floor))
		}
		staticParts = append(staticParts, fmt.Sprintf("%s=%s", role, RoleTier(role, models)))
	}
	line := fmt.Sprintf("routing (dynamic): unrouted [%s]", strings.Join(names, ", "))
	if len(floorParts) > 0 {
		line += "; floors: " + strings.Join(floorParts, ",")
	}
	line += "; static: " + strings.Join(staticParts, ",")
	return line + fmt.Sprintf("; decide: centinela route set %s <role> <tier> [--reason \"…\"]", feature)
}

// unroutedRoles preserves the caller's role order so the line is deterministic.
func unroutedRoles(roles []Role, routed map[string]string) []Role {
	var out []Role
	for _, role := range roles {
		if _, ok := routed[string(role)]; !ok {
			out = append(out, role)
		}
	}
	return out
}
