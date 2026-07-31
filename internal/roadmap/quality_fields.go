package roadmap

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// qualityScoresRaw is the pointer-typed mirror of QualityScores: a nil field
// after unmarshal means "absent", which a value-typed decode would have
// collapsed into a zero and mis-reported as an out-of-range score.
type qualityScoresRaw struct {
	AcceptanceCriteria *int `json:"acceptanceCriteria"`
	UserValue          *int `json:"userValue"`
	DefinitionClarity  *int `json:"definitionClarity"`
	Dependencies       *int `json:"dependencies"`
	EffortEstimation   *int `json:"effortEstimation"`
	Overall            *int `json:"overall"`
}

// scoreField pairs a score's JSON name with a pointer to its slot, so both the
// decoder and the range check can report WHICH field is at fault.
type scoreField struct {
	name string
	v    *int
}

// scoreFields and rawScoreFields are index-aligned by contract.
func scoreFields(s *QualityScores) []scoreField {
	return []scoreField{
		{"acceptanceCriteria", &s.AcceptanceCriteria},
		{"userValue", &s.UserValue},
		{"definitionClarity", &s.DefinitionClarity},
		{"dependencies", &s.Dependencies},
		{"effortEstimation", &s.EffortEstimation},
		{"overall", &s.Overall},
	}
}

func rawScoreFields(p *qualityScoresRaw) []scoreField {
	return []scoreField{
		{"acceptanceCriteria", p.AcceptanceCriteria},
		{"userValue", p.UserValue},
		{"definitionClarity", p.DefinitionClarity},
		{"dependencies", p.Dependencies},
		{"effortEstimation", p.EffortEstimation},
		{"overall", p.Overall},
	}
}

// collectScores copies the decoded pointers into a value struct, failing on the
// first absent field by name.
func collectScores(name string, p qualityScoresRaw) (QualityScores, error) {
	var s QualityScores
	dst := scoreFields(&s)
	for i, f := range rawScoreFields(&p) {
		if f.v == nil {
			return QualityScores{}, fmt.Errorf("feature %q: score field %q is missing — %s", name, f.name, scoreSchema)
		}
		*dst[i].v = *f.v
	}
	return s, nil
}

// scoreTypeError re-frames an encoding/json type error as a shape error naming
// the feature and the offending field — never as a range error.
func scoreTypeError(name string, err error) error {
	var te *json.UnmarshalTypeError
	if errors.As(err, &te) && te.Field != "" {
		return fmt.Errorf("feature %q: score field %q must be an integer 1-10, got JSON %s",
			name, te.Field, te.Value)
	}
	return fmt.Errorf(`feature %q: field "scores" is malformed (%v) — %s`, name, err, scoreSchema)
}

// jsonKind names the JSON kind of a raw value from its first token, so an error
// can state what was actually found.
func jsonKind(raw json.RawMessage) string {
	s := strings.TrimSpace(string(raw))
	if s == "" {
		return "null"
	}
	switch s[0] {
	case '{':
		return "object"
	case '[':
		return "array"
	case '"':
		return "string"
	case 't', 'f':
		return "boolean"
	case 'n':
		return "null"
	default:
		return "number"
	}
}
