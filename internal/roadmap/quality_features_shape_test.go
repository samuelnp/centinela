package roadmap

import (
	"strings"
	"testing"
)

// The features field itself is validated structurally, by name and expected shape.
func TestValidateQuality_FeaturesFieldFaultsAreStructural(t *testing.T) {
	cases := []struct{ name, body, want string }{
		{"absent", `{"role":"roadmap-quality-evaluator","threshold":9}`, `required field "features" is missing`},
		{"null", `{"role":"roadmap-quality-evaluator","threshold":9,"features":null}`, `required field "features" is missing`},
		{"object", `{"role":"roadmap-quality-evaluator","threshold":9,"features":{}}`, `"features" is a JSON object`},
		{"string", `{"role":"roadmap-quality-evaluator","threshold":9,"features":"none"}`, `"features" is a JSON string`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := qualityFixture(t, c.body)
			err := ValidateQuality(r)
			assertShapeError(t, err, c.want)
			if !strings.Contains(err.Error(), "array of {name, scores, summary}") {
				t.Fatalf("features error must state the expected shape, got %v", err)
			}
		})
	}
}

// The mirror direction: a genuinely out-of-range integer STILL produces the
// range error, and it now names the field and the value.
func TestValidateQuality_OutOfRangeStillReportsTheRange(t *testing.T) {
	cases := []struct{ name, features, field, value string }{
		{"too-high", `{"name":"user","scores":{"acceptanceCriteria":9,"userValue":11,"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9},"summary":"s"}`, "userValue", "11"},
		{"zero", `{"name":"user","scores":{"acceptanceCriteria":0,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9},"summary":"s"}`, "acceptanceCriteria", "0"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := qualityFixture(t, qualityReport(c.features))
			err := ValidateQuality(r)
			if err == nil || !strings.Contains(err.Error(), "must be between 1 and 10") {
				t.Fatalf("expected the range error, got %v", err)
			}
			if !strings.Contains(err.Error(), c.field) || !strings.Contains(err.Error(), c.value) {
				t.Fatalf("range error must name field %q and value %s, got %v", c.field, c.value, err)
			}
		})
	}
}

// Regression pin: a well-formed report still validates unchanged.
func TestValidateQuality_WellFormedReportStillValidates(t *testing.T) {
	r := qualityFixture(t, qualityReport(`{"name":"user",`+goodScores+`,"summary":"s"}`))
	if err := ValidateQuality(r); err != nil {
		t.Fatalf("valid report must still validate, got %v", err)
	}
}

// ParseScores keeps rejecting out-of-range CSV input and now names the field.
func TestParseScores_OutOfRangeNamesTheField(t *testing.T) {
	_, err := ParseScores("9,11,9,9,9,9")
	if err == nil || !strings.Contains(err.Error(), "must be between 1 and 10") {
		t.Fatalf("expected the range error, got %v", err)
	}
	if !strings.Contains(err.Error(), "userValue") {
		t.Fatalf("promote's range error must name the field, got %v", err)
	}
}
