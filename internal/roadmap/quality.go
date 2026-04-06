package roadmap

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const RoadmapQualityFile = ".workflow/roadmap-quality.json"
const RoadmapQualityMarkdown = ".workflow/roadmap-quality.md"
const qualityRole = "roadmap-quality-evaluator"
const qualityThreshold = 9

type QualityScores struct {
	AcceptanceCriteria int `json:"acceptanceCriteria"`
	UserValue          int `json:"userValue"`
	DefinitionClarity  int `json:"definitionClarity"`
	Dependencies       int `json:"dependencies"`
	EffortEstimation   int `json:"effortEstimation"`
	Overall            int `json:"overall"`
}

type QualityFeature struct {
	Name    string        `json:"name"`
	Scores  QualityScores `json:"scores"`
	Summary string        `json:"summary"`
}

type QualityReport struct {
	Role      string           `json:"role"`
	Threshold int              `json:"threshold"`
	Features  []QualityFeature `json:"features"`
}

func ValidateQuality(r *Roadmap) error {
	if _, err := os.Stat(RoadmapQualityMarkdown); err != nil {
		return fmt.Errorf("roadmap quality markdown missing: %s", RoadmapQualityMarkdown)
	}
	data, err := os.ReadFile(RoadmapQualityFile)
	if err != nil {
		return fmt.Errorf("roadmap quality json missing: %s", RoadmapQualityFile)
	}
	var q QualityReport
	if err := json.Unmarshal(data, &q); err != nil {
		return fmt.Errorf("invalid roadmap quality json: %w", err)
	}
	if q.Role != qualityRole {
		return fmt.Errorf("roadmap quality role must be %s", qualityRole)
	}
	if q.Threshold != qualityThreshold {
		return fmt.Errorf("roadmap quality threshold must be %d", qualityThreshold)
	}
	names := roadmapFeatureSet(r)
	seen := map[string]bool{}
	for _, f := range q.Features {
		if !names[f.Name] {
			return fmt.Errorf("quality references unknown feature: %s", f.Name)
		}
		if err := validateScoreRange(f.Scores); err != nil {
			return fmt.Errorf("feature %s has invalid scores: %w", f.Name, err)
		}
		if f.Scores.Overall < qualityThreshold {
			return fmt.Errorf("feature %s overall score %d is below %d", f.Name, f.Scores.Overall, qualityThreshold)
		}
		if strings.TrimSpace(f.Summary) == "" {
			return fmt.Errorf("feature %s summary is required", f.Name)
		}
		seen[f.Name] = true
	}
	for name := range names {
		if !seen[name] {
			return fmt.Errorf("quality missing feature: %s", name)
		}
	}
	return nil
}

func validateScoreRange(s QualityScores) error {
	vals := []int{s.AcceptanceCriteria, s.UserValue, s.DefinitionClarity, s.Dependencies, s.EffortEstimation, s.Overall}
	for _, v := range vals {
		if v < 1 || v > 10 {
			return fmt.Errorf("scores must be between 1 and 10")
		}
	}
	return nil
}
