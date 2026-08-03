package roadmap

import (
	"encoding/json"
	"fmt"
	"os"
)

// RoadmapAnalysisFile is the machine-readable senior-PM roadmap analysis.
const RoadmapAnalysisFile = ".workflow/roadmap-analysis.json"

// RoadmapAnalysisMarkdown is its human-readable companion; both must exist.
const RoadmapAnalysisMarkdown = ".workflow/roadmap-analysis.md"

// AnalysisFeature is a single feature entry in the senior-PM analysis.
// DependsOn was removed in Option B — dependencies are now on roadmap.json.
type AnalysisFeature struct {
	Name string `json:"name"`
}

// Analysis is the decoded senior-PM roadmap analysis. Provisional marks an
// artifact seeded by `roadmap promote` rather than written by an evaluator; see
// provisional.go for why the role string alone is not enough.
type Analysis struct {
	Role        string            `json:"role"`
	Provisional bool              `json:"provisional,omitempty"`
	Features    []AnalysisFeature `json:"features"`
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
	if a.Provisional {
		return provisionalRefusal(RoadmapAnalysisFile, "senior-PM analysis")
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

// roadmapFeatureSet returns the features that require analysis/quality
// coverage. Backlog-phase findings are exempt — they carry no analysis or
// quality entries until promoted. The exemption is defined once, in
// NonBacklogFeatureSet, and shared by ValidateAnalysis and ValidateQuality.
func roadmapFeatureSet(r *Roadmap) map[string]bool {
	return NonBacklogFeatureSet(r)
}
