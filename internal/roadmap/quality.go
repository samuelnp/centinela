package roadmap

import (
	"encoding/json"
	"fmt"
	"os"
)

// RoadmapQualityFile is the machine-readable roadmap quality report.
const RoadmapQualityFile = ".workflow/roadmap-quality.json"

// RoadmapQualityMarkdown is its human-readable companion; both must exist.
const RoadmapQualityMarkdown = ".workflow/roadmap-quality.md"
const qualityRole = "roadmap-quality-evaluator"

// QualityScores is one feature's six evaluator dimensions, each an integer
// 1-10. The scores are recorded and range-checked; none of them gates anything.
type QualityScores struct {
	AcceptanceCriteria int `json:"acceptanceCriteria"`
	UserValue          int `json:"userValue"`
	DefinitionClarity  int `json:"definitionClarity"`
	Dependencies       int `json:"dependencies"`
	EffortEstimation   int `json:"effortEstimation"`
	Overall            int `json:"overall"`
}

// QualityFeature is one graded roadmap feature: its name, scores and the
// evaluator's required summary.
type QualityFeature struct {
	Name    string        `json:"name"`
	Scores  QualityScores `json:"scores"`
	Summary string        `json:"summary"`
}

// QualityReport is the decoded roadmap quality artifact. Threshold is retained
// for backward compatibility with reports written before the minimum-score gate
// was deleted; it is never enforced.
type QualityReport struct {
	Role      string           `json:"role"`
	Threshold int              `json:"threshold"`
	Features  []QualityFeature `json:"features"`
}

// qualityEnvelope is the structural first stage: role is decoded eagerly (it
// gates everything else), while features stays raw so a shape fault inside it
// can be named as a shape fault rather than decoding to zeros. Threshold still
// DECODES so every artifact written before the threshold gate was deleted keeps
// parsing — its value is deliberately ignored, whatever it says.
type qualityEnvelope struct {
	Role      string           `json:"role"`
	Threshold int              `json:"threshold"`
	Features  *json.RawMessage `json:"features"`
}

// ValidateQuality checks the roadmap quality artifacts: both files present, the
// evaluator role correct, every feature structurally sound with in-range scores
// and a summary, and roadmap/report feature sets in agreement. It enforces no
// score MINIMUM: a number the evaluated party assigns itself was never evidence,
// so the old overall >= 9 refusal is gone in every profile (see QualityAdvisories).
func ValidateQuality(r *Roadmap) error {
	features, err := loadQualityFeatures()
	if err != nil {
		return err
	}
	return validateQualityFeatures(r, features)
}

// loadQualityFeatures reads and structurally decodes the quality report,
// stopping at the first file, role or shape fault. Shared by ValidateQuality
// and the advisory reporter so both see one decoding path.
func loadQualityFeatures() ([]QualityFeature, error) {
	if _, err := os.Stat(RoadmapQualityMarkdown); err != nil {
		return nil, fmt.Errorf("roadmap quality markdown missing: %s", RoadmapQualityMarkdown)
	}
	data, err := os.ReadFile(RoadmapQualityFile)
	if err != nil {
		return nil, fmt.Errorf("roadmap quality json missing: %s", RoadmapQualityFile)
	}
	var q qualityEnvelope
	if err := json.Unmarshal(data, &q); err != nil {
		return nil, fmt.Errorf("invalid roadmap quality json: %w", err)
	}
	if q.Role != qualityRole {
		return nil, fmt.Errorf("roadmap quality role must be %s", qualityRole)
	}
	return decodeQualityFeatures(q.Features)
}
