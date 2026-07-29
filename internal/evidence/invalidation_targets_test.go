package evidence

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func hasRole(roles []Role, r string) bool {
	for _, x := range roles {
		if string(x) == r {
			return true
		}
	}
	return false
}

func TestInvalidationTargetsValidate(t *testing.T) {
	setupWF(t)
	roles, artifacts := InvalidationTargets("f", "validate")
	// gatekeeper now arrives via RequiredRolesForFeature (once, not twice);
	// validation-specialist is shed as a legacy certification.
	if !hasRole(roles, "validation-specialist") ||
		!hasRole(roles, "gatekeeper") || !hasRole(roles, "production-readiness") {
		t.Fatalf("validate roles = %v", roles)
	}
	if len(artifacts) != 0 {
		t.Fatalf("validate artifacts = %v", artifacts)
	}
}

func TestInvalidationTargetsTests(t *testing.T) {
	setupWF(t)
	roles, artifacts := InvalidationTargets("f", "tests")
	if !hasRole(roles, "qa-senior") {
		t.Fatalf("tests roles = %v", roles)
	}
	if len(artifacts) != 1 || artifacts[0] != "edge-cases.md" {
		t.Fatalf("tests artifacts = %v", artifacts)
	}
}

func TestInvalidationTargetsCodeInternalVsUserFacing(t *testing.T) {
	setupWF(t)
	// Internal feature: no docs/features file → ux-ui-specialist excluded.
	roles, _ := InvalidationTargets("internal-feat", "code")
	if hasRole(roles, "ux-ui-specialist") {
		t.Fatalf("internal must exclude ux-ui: %v", roles)
	}
	if !hasRole(roles, "senior-engineer") {
		t.Fatalf("code must include senior-engineer: %v", roles)
	}
	// User-facing feature: surface marker present → ux-ui-specialist included.
	os.MkdirAll("docs/features", 0o755)                                             //nolint:errcheck
	os.WriteFile("docs/features/ufeat.md", []byte("Surface: user-facing\n"), 0o644) //nolint:errcheck
	roles, _ = InvalidationTargets("ufeat", "code")
	if !hasRole(roles, "ux-ui-specialist") {
		t.Fatalf("user-facing must include ux-ui: %v", roles)
	}
}

// InvalidationTargets has no "plan" branch and is never called with "plan":
// revise is backward-only, so reopenedSteps (order[idx+1:]) can never yield the
// index-0 step. This asserts the CURRENT behavior — planner arrives from
// RequiredRolesForFeature and nothing extra is added — and deliberately does NOT
// assert that a rewind to plan sheds anything, because it does not. That gap is
// deferred as `revise-to-plan-sheds-no-evidence`.
func TestInvalidationTargetsPlanNoExtras(t *testing.T) {
	setupWF(t)
	roles, artifacts := InvalidationTargets("f", "plan")
	if len(artifacts) != 0 {
		t.Fatalf("plan artifacts = %v", artifacts)
	}
	if !hasRole(roles, "planner") {
		t.Fatalf("plan roles must come from RequiredRolesForFeature: %v", roles)
	}
	if hasRole(roles, "big-thinker") || hasRole(roles, "feature-specialist") {
		t.Fatalf("no retired-role branch exists for plan; got %v", roles)
	}
}

// Guard for the dead-branch trap itself: plan must be index 0 of every step
// order, which is WHY the function is unreachable for "plan". If a future step
// order ever puts something before plan, this fails and the deferred fix
// becomes live work rather than a silent no-op.
func TestPlanIsFirstStepSoInvalidationNeverSeesIt(t *testing.T) {
	for _, order := range [][]string{workflow.DefaultStepOrder, workflow.BootstrapStepOrder} {
		if len(order) == 0 || order[0] != "plan" {
			t.Fatalf("plan must be the first step for the no-plan-branch reasoning to hold: %v", order)
		}
	}
}
