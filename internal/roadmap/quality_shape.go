package roadmap

import (
	"encoding/json"
	"fmt"
)

// scoreSchema is the sentence every scores shape fault ends with, so the
// operator is told the schema that was expected instead of hunting for an
// out-of-range number that does not exist.
const scoreSchema = `expected an object {acceptanceCriteria, userValue, ` +
	`definitionClarity, dependencies, effortEstimation, overall}, ` +
	`each an integer 1-10`

const featuresSchema = `expected an array of {name, scores, summary} objects`

// qualityFeatureRaw is the structural mirror of QualityFeature: scores stays a
// raw message so "absent", "wrong JSON kind" and "wrong field type" are three
// distinguishable faults instead of one silent decode-to-zero.
type qualityFeatureRaw struct {
	Name    string           `json:"name"`
	Scores  *json.RawMessage `json:"scores"`
	Summary string           `json:"summary"`
}

// decodeQualityFeatures decodes the features array in two stages: structural
// first (is the field present, is it the right JSON kind), per-field second.
// A shape fault is NEVER reported as a range fault.
func decodeQualityFeatures(raw *json.RawMessage) ([]QualityFeature, error) {
	if raw == nil || jsonKind(*raw) == "null" {
		return nil, fmt.Errorf(`roadmap quality json: required field "features" is missing — %s`, featuresSchema)
	}
	if kind := jsonKind(*raw); kind != "array" {
		return nil, fmt.Errorf(`roadmap quality json: field "features" is a JSON %s — %s`, kind, featuresSchema)
	}
	var rawFeatures []qualityFeatureRaw
	if err := json.Unmarshal(*raw, &rawFeatures); err != nil {
		return nil, fmt.Errorf(`roadmap quality json: field "features" is malformed (%v) — %s`, err, featuresSchema)
	}
	out := make([]QualityFeature, 0, len(rawFeatures))
	for _, rf := range rawFeatures {
		scores, err := decodeQualityScores(rf.Name, rf.Scores)
		if err != nil {
			return nil, err
		}
		out = append(out, QualityFeature{Name: rf.Name, Scores: scores, Summary: rf.Summary})
	}
	return out, nil
}

// decodeQualityScores turns one feature's raw scores value into QualityScores,
// naming the feature and the offending field for every structural fault.
func decodeQualityScores(name string, raw *json.RawMessage) (QualityScores, error) {
	if raw == nil || jsonKind(*raw) == "null" {
		return QualityScores{}, fmt.Errorf(`feature %q: required object field "scores" is missing — %s`, name, scoreSchema)
	}
	if kind := jsonKind(*raw); kind != "object" {
		return QualityScores{}, fmt.Errorf(`feature %q: field "scores" is a JSON %s — %s`, name, kind, scoreSchema)
	}
	var p qualityScoresRaw
	if err := json.Unmarshal(*raw, &p); err != nil {
		return QualityScores{}, scoreTypeError(name, err)
	}
	return collectScores(name, p)
}
