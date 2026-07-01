package evidence

import (
	"os"
	"testing"
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

func TestInvalidationTargetsPlanNoExtras(t *testing.T) {
	setupWF(t)
	roles, artifacts := InvalidationTargets("f", "plan")
	if len(artifacts) != 0 {
		t.Fatalf("plan artifacts = %v", artifacts)
	}
	if !hasRole(roles, "big-thinker") || !hasRole(roles, "feature-specialist") {
		t.Fatalf("plan roles = %v", roles)
	}
}
