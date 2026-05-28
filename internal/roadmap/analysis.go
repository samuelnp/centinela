package roadmap

import (
	"encoding/json"
	"fmt"
	"os"
)

const RoadmapAnalysisFile = ".workflow/roadmap-analysis.json"
const RoadmapAnalysisMarkdown = ".workflow/roadmap-analysis.md"

// AnalysisFeature is a single feature entry in the senior-PM analysis.
// DependsOn was removed in Option B — dependencies are now on roadmap.json.
type AnalysisFeature struct {
	Name string `json:"name"`
}

type Analysis struct {
	Role     string            `json:"role"`
	Features []AnalysisFeature `json:"features"`
}

// ValidateAnalysis checks role, markdown presence, and feature coverage only.
// Cycle and unknown-dep validation are now handled by ValidateDependencies.
func ValidateAnalysis(r *Roadmap) error {
	if _, err := os.Stat(RoadmapAnalysisMarkdown); err != nil {
		return fmt.Errorf("roadmap analysis markdown missing: %s", RoadmapAnalysisMarkdown)
	}
	data, err := os.ReadFile(RoadmapAnalysisFile)
	if err != nil {
		return fmt.Errorf("roadmap analysis json missing: %s", RoadmapAnalysisFile)
	}
	var a Analysis
	if err := json.Unmarshal(data, &a); err != nil {
		return fmt.Errorf("invalid roadmap analysis json: %w", err)
	}
	if a.Role != "senior-product-manager" {
		return fmt.Errorf("roadmap analysis role must be senior-product-manager")
	}
	names := roadmapFeatureSet(r)
	seen := map[string]bool{}
	for _, f := range a.Features {
		if !names[f.Name] {
			return fmt.Errorf("analysis references unknown feature: %s", f.Name)
		}
		seen[f.Name] = true
	}
	for name := range names {
		if !seen[name] {
			return fmt.Errorf("analysis missing feature: %s", name)
		}
	}
	return nil
}

func roadmapFeatureSet(r *Roadmap) map[string]bool {
	out := map[string]bool{}
	if r == nil {
		return out
	}
	for _, p := range r.Phases {
		for _, f := range p.Features {
			out[f.Name] = true
		}
	}
	return out
}
