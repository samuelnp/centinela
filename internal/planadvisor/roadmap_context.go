package planadvisor

import (
	"github.com/samuelnp/centinela/internal/roadmap"
)

func relatedNames(feature string) ([]string, []string) {
	deps := take(dependencyNames(feature), 2)
	return deps, take(siblingNames(feature, deps), 2)
}

// dependencyNames reads dependsOn for the given feature from roadmap.json.
// Returns nil when roadmap is missing or the feature has no deps.
func dependencyNames(feature string) []string {
	r, err := roadmap.Load()
	if err != nil {
		return nil
	}
	for _, phase := range r.Phases {
		for _, f := range phase.Features {
			if f.Name == feature {
				return append([]string{}, f.DependsOn...)
			}
		}
	}
	return nil
}

func siblingNames(feature string, deps []string) []string {
	r, err := roadmap.Load()
	if err != nil {
		return nil
	}
	seen := map[string]bool{feature: true}
	for _, dep := range deps {
		seen[dep] = true
	}
	for _, phase := range r.Phases {
		if !phaseHasFeature(phase, feature) {
			continue
		}
		out := []string{}
		for _, f := range phase.Features {
			if !seen[f.Name] {
				out = append(out, f.Name)
			}
		}
		return out
	}
	return nil
}

func phaseHasFeature(p roadmap.Phase, feature string) bool {
	for _, f := range p.Features {
		if f.Name == feature {
			return true
		}
	}
	return false
}
