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
		// gatekeeper arrives via RequiredRolesForFeature since adversarial-v1.
		// production-readiness is out-of-band, and validation-specialist is
		// shed even though no step requires it any more: a legacy in-flight
		// workflow that rewinds past validate must not keep a stale legacy
		// pair that would satisfy its legacy gate on the next pass.
		roles = append(roles, Role("production-readiness"), orchestration.RoleValidationSpec)
	case "plan":
		// planner arrives via RequiredRolesForFeature since planner-v1. The two
		// retired roles are shed as well, for the same reason validate sheds
		// validation-specialist: a legacy in-flight workflow that rewinds past
		// plan must not keep a stale legacy pair that would immediately satisfy
		// its legacy gate on the next pass.
		roles = append(roles, orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial)
	case "tests":
		artifacts = append(artifacts, "edge-cases.md")
	}
	return roles, artifacts
}
