package roadmap

import (
	"encoding/json"
	"fmt"
	"os"
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

// qualityEnvelope is the structural first stage: role and threshold are decoded
// eagerly (they gate everything else), while features stays raw so a shape
// fault inside it can be named as a shape fault rather than decoding to zeros.
type qualityEnvelope struct {
	Role      string           `json:"role"`
	Threshold int              `json:"threshold"`
	Features  *json.RawMessage `json:"features"`
}

func ValidateQuality(r *Roadmap) error {
	if _, err := os.Stat(RoadmapQualityMarkdown); err != nil {
		return fmt.Errorf("roadmap quality markdown missing: %s", RoadmapQualityMarkdown)
	}
	data, err := os.ReadFile(RoadmapQualityFile)
	if err != nil {
		return fmt.Errorf("roadmap quality json missing: %s", RoadmapQualityFile)
	}
	var q qualityEnvelope
	if err := json.Unmarshal(data, &q); err != nil {
		return fmt.Errorf("invalid roadmap quality json: %w", err)
	}
	if q.Role != qualityRole {
		return fmt.Errorf("roadmap quality role must be %s", qualityRole)
	}
	if q.Threshold != qualityThreshold {
		return fmt.Errorf("roadmap quality threshold must be %d", qualityThreshold)
	}
	features, err := decodeQualityFeatures(q.Features)
	if err != nil {
		return err
	}
	return validateQualityFeatures(r, features)
}
