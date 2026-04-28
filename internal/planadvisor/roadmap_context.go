package planadvisor

import (
	"encoding/json"
	"os"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func relatedNames(feature string) ([]string, []string) {
	deps := take(dependencyNames(feature), 2)
	return deps, take(siblingNames(feature, deps), 2)
}

func dependencyNames(feature string) []string {
	data, err := os.ReadFile(roadmap.RoadmapAnalysisFile)
	if err != nil {
		return nil
	}
	var a roadmap.Analysis
	if json.Unmarshal(data, &a) != nil {
		return nil
	}
	for _, f := range a.Features {
		if f.Name == feature {
			return append([]string{}, f.DependsOn...)
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
