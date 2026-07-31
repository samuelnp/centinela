// Acceptance: specs/truthful-validators.feature
//
// Section E (part 2) — a genuinely out-of-range score still produces the
// range error, a missing/non-array features list is a structural error, and a
// well-formed report still validates unchanged.
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: A genuinely out-of-range score still produces the range error
func TestTV_Quality_OutOfRangeStillReportsTheRange(t *testing.T) {
	scores := `{"acceptanceCriteria":9,"userValue":11,"definitionClarity":9,` +
		`"dependencies":9,"effortEstimation":9,"overall":9}`
	q := `{"role":"roadmap-quality-evaluator","threshold":9,` +
		`"features":[{"name":"alpha","scores":` + scores + `,"summary":"s"}]}`
	out, code := tvQualityFixture(t, q)
	if code == 0 {
		t.Fatalf("an out-of-range score must fail validate\n%s", out)
	}
	mustContain(t, out, "between 1 and 10")
	mustContain(t, out, "userValue")
	mustContain(t, out, "11")
}

// Scenario: A missing or non-array features list is reported structurally
func TestTV_Quality_MissingFeaturesListIsStructural(t *testing.T) {
	q := `{"role":"roadmap-quality-evaluator","threshold":9}`
	out, code := tvQualityFixture(t, q)
	if code == 0 {
		t.Fatalf("a missing features list must fail validate\n%s", out)
	}
	mustContain(t, out, "features")
	if strings.Contains(out, "between 1 and 10") {
		t.Fatalf("a missing features list must not be reported as a range fault\n%s", out)
	}
}

// Scenario: A well-formed quality report still validates unchanged
func TestTV_Quality_WellFormedReportStillValidates(t *testing.T) {
	q := `{"role":"roadmap-quality-evaluator","threshold":9,` +
		`"features":[{"name":"alpha","scores":` + tvFullScores + `,"summary":"s"}]}`
	out, code := tvQualityFixture(t, q)
	if code != 0 {
		t.Fatalf("a well-formed quality report must validate\n%s", out)
	}
	mustContain(t, out, "valid")
}
