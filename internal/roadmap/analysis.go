package roadmap

import (
	"encoding/json"
	"fmt"
	"os"
)

const RoadmapAnalysisFile = ".workflow/roadmap-analysis.json"
const RoadmapAnalysisMarkdown = ".workflow/roadmap-analysis.md"

type AnalysisFeature struct {
	Name      string   `json:"name"`
	DependsOn []string `json:"dependsOn"`
}

type Analysis struct {
	Role     string            `json:"role"`
	Features []AnalysisFeature `json:"features"`
}

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
	deps := map[string][]string{}
	for _, f := range a.Features {
		if !names[f.Name] {
			return fmt.Errorf("analysis references unknown feature: %s", f.Name)
		}
		seen[f.Name] = true
		for _, dep := range f.DependsOn {
			if !names[dep] {
				return fmt.Errorf("feature %s depends on unknown feature %s", f.Name, dep)
			}
		}
		deps[f.Name] = f.DependsOn
	}
	for name := range names {
		if !seen[name] {
			return fmt.Errorf("analysis missing feature: %s", name)
		}
	}
	if hasCycle(deps) {
		return fmt.Errorf("roadmap analysis contains dependency cycle")
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
