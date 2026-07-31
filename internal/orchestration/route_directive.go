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
	unrouted := unroutedRoles(roles, routed, floors)
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
		// "(routes only)" is load-bearing honesty: a floor bounds what `route
		// set` may record and what the overlay will honor — it does NOT raise a
		// static tier the project chose in [orchestration.models]. Printing a
		// bare "floors:" beside a lower "static:" claimed an enforcement the
		// static path deliberately does not perform.
		line += "; floors (routes only): " + strings.Join(floorParts, ",")
	}
	line += "; static: " + strings.Join(staticParts, ",")
	return line + fmt.Sprintf("; decide: centinela route set %s <role> <tier> [--reason \"…\"]", feature)
}

// unroutedRoles preserves the caller's role order so the line is deterministic.
// A role whose recorded route is NOT honored (corrupt tier, or below its floor)
// counts as un-routed: the overlay ignores it, so the decision is still open and
// the operator must see it rather than read silence as agreement.
func unroutedRoles(roles []Role, routed, floors map[string]string) []Role {
	var out []Role
	for _, role := range roles {
		if _, honored := HonoredRouteTier(role, routed[string(role)], floors); !honored {
			out = append(out, role)
		}
	}
	return out
}
