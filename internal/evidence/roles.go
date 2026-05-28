// Package evidence provides typed CLI authoring + validation for the per-role
// .workflow/<feature>-<role>.json artifacts described in
// docs/architecture/evidence-contract.md. All schema rules are delegated to
// the existing validator in internal/orchestration so there is only one
// source of truth.
package evidence

import (
	"fmt"
	"sort"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// Role is the slug an agent identifies itself as in the evidence file. It is
// an alias-ish wrapper around orchestration.Role so internal/evidence does
// not import cmd/ and can be referenced from cmd/.
type Role = orchestration.Role

// AllRoles returns every role the CLI supports, in stable order. Documented
// roles in evidence-contract.md plus merge-steward (out-of-band).
func AllRoles() []Role {
	return []Role{
		orchestration.RoleBigThinker,
		orchestration.RoleFeatureSpecial,
		orchestration.RoleSeniorEngineer,
		orchestration.RoleUXUISpecialist,
		orchestration.RoleQASeniorEngineer,
		orchestration.RoleValidationSpec,
		orchestration.RoleDocsSpecialist,
		orchestration.RoleMergeSteward,
		Role("gatekeeper"),
		Role("production-readiness"),
	}
}

// IsKnownRole reports whether r appears in AllRoles().
func IsKnownRole(r Role) bool {
	for _, k := range AllRoles() {
		if k == r {
			return true
		}
	}
	return false
}

// ParseRole returns the matching Role or an error listing the accepted slugs.
func ParseRole(s string) (Role, error) {
	r := Role(s)
	if IsKnownRole(r) {
		return r, nil
	}
	all := make([]string, 0, len(AllRoles()))
	for _, k := range AllRoles() {
		all = append(all, string(k))
	}
	sort.Strings(all)
	return "", fmt.Errorf("unknown role %q (allowed: %v)", s, all)
}
