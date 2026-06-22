package reconstruct

import (
	"sort"

	"github.com/samuelnp/centinela/internal/analyze"
)

// Select builds the deterministic, sorted, de-duplicated Target set from the
// Inventory: it applies the exclude table (exclusion wins), then the first
// matching promote rule, slugifies each survivor, disambiguates slug collisions,
// sorts by slug ascending, and bounds the result to maxTargets. The original
// inv.Packages order is preserved through promotion so disambiguation suffixes
// are assigned stably.
func Select(inv analyze.Inventory) []Target {
	s := newSignals(inv)
	var targets []Target
	for i, pkg := range inv.Packages {
		low := s.pkgs[i]
		if excluded(low) {
			continue
		}
		role, reason, ok := promote(low, s)
		if !ok {
			continue
		}
		targets = append(targets, Target{Pkg: pkg, Role: role, Reason: reason})
	}
	assignSlugs(targets)
	sort.Slice(targets, func(i, j int) bool { return targets[i].Slug < targets[j].Slug })
	if len(targets) > maxTargets {
		targets = targets[:maxTargets]
	}
	return targets
}

func excluded(lowPkg string) bool {
	for _, r := range excludeRules {
		if r.match(lowPkg) {
			return true
		}
	}
	return false
}

func promote(lowPkg string, s signals) (Role, string, bool) {
	for _, r := range promoteRules {
		if r.match(lowPkg, s) {
			return r.role, r.reason, true
		}
	}
	return "", "", false
}

// assignSlugs fills each target's Slug, disambiguating collisions deterministically
// in the packages' original (pre-sort) order.
func assignSlugs(targets []Target) {
	used := map[string]bool{}
	for i := range targets {
		targets[i].Slug = disambiguate(slugify(targets[i].Pkg), used)
	}
}
