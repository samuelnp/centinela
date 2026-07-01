package evidence

import "github.com/samuelnp/centinela/internal/orchestration"

// InvalidationTargets returns the certification artifacts a single re-opened
// step must shed so the next `centinela complete` re-runs its gates: the step's
// required roles (feature-aware — ux-ui-specialist only for user-facing code,
// documentation-specialist skipped for internal docs), PLUS the validate step's
// out-of-band gatekeeper + production-readiness reports, PLUS the tests step's
// non-role -edge-cases.md artifact. This is the one new piece of revise policy;
// it lives here (not cmd/) so it stays unit-testable and G7-clean. The caller
// dedupes across steps. artifacts entries are path suffixes (with extension).
func InvalidationTargets(feature, step string) (roles []Role, artifacts []string) {
	roles = append(roles, orchestration.RequiredRolesForFeature(feature, step)...)
	switch step {
	case "validate":
		roles = append(roles, Role("gatekeeper"), Role("production-readiness"))
	case "tests":
		artifacts = append(artifacts, "edge-cases.md")
	}
	return roles, artifacts
}
