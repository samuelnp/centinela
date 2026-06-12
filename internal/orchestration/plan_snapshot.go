package orchestration

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

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

func requiresPlanSnapshot(role Role) bool {
	return role == RoleBigThinker || role == RoleFeatureSpecial
}

// RequiredPlanInputs returns the plan-snapshot input set the big-thinker and
// feature-specialist must list: the feature's own brief, every docs/features/*.md,
// and the feature's plan at docs/plans/<feature>.md — normalized and sorted (the
// set the evidence contract documents as required). The validator and the
// evidence init pre-fill share this so a pre-filled init validates by construction.
func RequiredPlanInputs(feature string) []string {
	seen := map[string]struct{}{}
	required := []string{
		normalizeFeatureDocPath("docs/features/" + feature + ".md"),
		normalizeFeatureDocPath("docs/plans/" + feature + ".md"),
	}
	for _, p := range required {
		seen[p] = struct{}{}
	}
	files, _ := filepath.Glob("docs/features/*.md")
	for _, f := range files {
		n := normalizeFeatureDocPath(f)
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		required = append(required, n)
	}
	sort.Strings(required)
	return required
}

func normalizeFeatureDocPath(p string) string {
	n := strings.ReplaceAll(strings.TrimSpace(p), "\\", "/")
	i := strings.Index(n, "docs/features/")
	if i >= 0 {
		n = n[i:]
	}
	n = strings.TrimPrefix(n, "./")
	return filepath.ToSlash(filepath.Clean(n))
}
