package orchestration

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// validatePlanSnapshotInputs reports only the MISSING required entries. The
// rule is include, never equal: evidence that lists additional inputs — the
// shape every workflow planned under the old repo-wide glob already wrote —
// stays valid. Do not tidy this into a set-equality check.
func validatePlanSnapshotInputs(path, feature, step string, role Role, inputs []string) error {
	if step != "plan" || !requiresPlanSnapshot(role) {
		return nil
	}
	required := RequiredPlanInputs(feature)
	provided := map[string]struct{}{}
	for _, in := range inputs {
		provided[normalizeFeatureDocPath(in)] = struct{}{}
	}
	missing := []string{}
	for _, want := range required {
		if _, ok := provided[want]; !ok {
			missing = append(missing, want)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("missing feature-doc snapshot inputs: %s in: %s", strings.Join(missing, ", "), path)
}

// requiresPlanSnapshot covers the planner plus the two retired legacy plan
// roles, so an in-flight legacy workflow gets the same shrunken required set;
// its already-written repo-wide evidence still passes as a superset.
func requiresPlanSnapshot(role Role) bool {
	return role == RolePlanner || role == RoleBigThinker || role == RoleFeatureSpecial
}

// RequiredPlanInputs returns the plan-snapshot input set every plan role must
// list: exactly the feature's OWN brief (docs/features/<feature>.md) and its
// OWN plan (docs/plans/<feature>.md), normalized and sorted. It is derived by
// construction — it performs no filesystem glob, so the set never grows with
// the number of briefs in the repo and is identical in an empty checkout.
// The validator and the evidence init pre-fill share this so a pre-filled init
// validates by construction. Callers may list MORE inputs; the rule is include.
func RequiredPlanInputs(feature string) []string {
	required := []string{
		normalizeFeatureDocPath("docs/features/" + feature + ".md"),
		normalizeFeatureDocPath("docs/plans/" + feature + ".md"),
	}
	sort.Strings(required)
	return required
}

// normalizeFeatureDocPath anchors a written input on the first occurrence of
// either docs/features/ or docs/plans/ so an absolute or deeply-prefixed path
// still matches, symmetrically for both required entries. Backslashes are
// folded to slashes first so a Windows-style path anchors too.
func normalizeFeatureDocPath(p string) string {
	n := strings.ReplaceAll(strings.TrimSpace(p), "\\", "/")
	if i := docAnchorIndex(n); i >= 0 {
		n = n[i:]
	}
	n = strings.TrimPrefix(n, "./")
	return filepath.ToSlash(filepath.Clean(n))
}

// docAnchorIndex returns the earliest index at which a known plan-doc directory
// prefix occurs, or -1 when the path carries neither.
func docAnchorIndex(n string) int {
	best := -1
	for _, anchor := range []string{"docs/features/", "docs/plans/"} {
		if i := strings.Index(n, anchor); i >= 0 && (best < 0 || i < best) {
			best = i
		}
	}
	return best
}
