package roadmap

import (
	"os"
	"strings"
	"testing"
)

// qualityFixture chdirs into a temp repo carrying the quality markdown and the
// given raw JSON body, and returns the roadmap the report is validated against.
func qualityFixture(t *testing.T, body string) *Roadmap {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(RoadmapQualityMarkdown, []byte("# q"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(RoadmapQualityFile, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{{Name: "user"}}}}}
}

// A shape fault must name the shape and must never borrow the range message.
func TestValidateQuality_ScoresShapeFaultsNameTheShape(t *testing.T) {
	cases := []struct{ name, features, want string }{
		{"missing", `{"name":"user","summary":"s"}`, `required object field "scores" is missing`},
		{"null", `{"name":"user","scores":null,"summary":"s"}`, `required object field "scores" is missing`},
		{"array", `{"name":"user","scores":[9,9,9,9,9,9],"summary":"s"}`, `"scores" is a JSON array`},
		{"string", `{"name":"user","scores":"9","summary":"s"}`, `"scores" is a JSON string`},
		{"number", `{"name":"user","scores":9,"summary":"s"}`, `"scores" is a JSON number`},
		{"boolean", `{"name":"user","scores":true,"summary":"s"}`, `"scores" is a JSON boolean`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := qualityFixture(t, qualityReport(c.features))
			err := ValidateQuality(r)
			assertShapeError(t, err, c.want)
			if !strings.Contains(err.Error(), "acceptanceCriteria") {
				t.Fatalf("shape error must list the expected fields, got %v", err)
			}
		})
	}
}

// A per-field fault names that field — a missing one, a float, a numeric string.
func TestValidateQuality_ScoreFieldFaultsNameTheField(t *testing.T) {
	partial := `"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,"dependencies":9`
	cases := []struct{ name, features, want string }{
		{"field-missing", `{"name":"user","scores":{` + partial + `,"overall":9},"summary":"s"}`, "effortEstimation"},
		{"float", `{"name":"user","scores":{` + partial + `,"effortEstimation":9,"overall":9.0},"summary":"s"}`, "overall"},
		{"numeric-string", `{"name":"user","scores":{` + partial + `,"effortEstimation":9,"overall":"9"},"summary":"s"}`, "overall"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := qualityFixture(t, qualityReport(c.features))
			assertShapeError(t, ValidateQuality(r), c.want)
		})
	}
}

// assertShapeError pins both directions: the message says what is wrong AND it
// is not the range message, which is reserved for a real out-of-range integer.
func assertShapeError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected a shape error mentioning %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("want %q in error, got %v", want, err)
	}
	if strings.Contains(err.Error(), "between 1 and 10") {
		t.Fatalf("a shape fault must not be reported as a range fault: %v", err)
	}
}
