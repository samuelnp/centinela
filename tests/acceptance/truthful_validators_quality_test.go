// Acceptance: specs/truthful-validators.feature
//
// Section E (part 1) — roadmap-quality shape errors name the shape instead of
// reporting a range fault for a structural problem. Every case shares a valid
// roadmap + analysis pair so ValidateAnalysis (which runs first) always
// passes and only ValidateQuality's verdict is under test.
package acceptance_test

import (
	"strings"
	"testing"
)

const tvRoadmapJSON = `{"phases":[{"name":"Phase 0","features":[{"name":"alpha"}]}]}`
const tvAnalysisJSON = `{"role":"senior-product-manager","features":[{"name":"alpha"}]}`

// tvQualityFixture writes a scratch repo with a valid roadmap + analysis pair
// and the given raw roadmap-quality.json body, then runs `roadmap validate`.
func tvQualityFixture(t *testing.T, qualityJSON string) (string, int) {
	t.Helper()
	bin := buildCent(t)
	dir := t.TempDir()
	writeFile(t, dir, ".workflow/roadmap.json", tvRoadmapJSON)
	writeFile(t, dir, ".workflow/roadmap-analysis.json", tvAnalysisJSON)
	writeFile(t, dir, ".workflow/roadmap-analysis.md", "# a\n")
	writeFile(t, dir, ".workflow/roadmap-quality.json", qualityJSON)
	writeFile(t, dir, ".workflow/roadmap-quality.md", "# q\n")
	return runCent(t, bin, dir, "roadmap", "validate")
}

const tvFullScores = `{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,` +
	`"dependencies":9,"effortEstimation":9,"overall":9}`

// Scenario: A missing scores object is reported as a missing object, not a bad number
func TestTV_Quality_MissingScoresIsAShapeError(t *testing.T) {
	q := `{"role":"roadmap-quality-evaluator","threshold":9,` +
		`"features":[{"name":"alpha","summary":"s"}]}`
	out, code := tvQualityFixture(t, q)
	if code == 0 {
		t.Fatalf("a missing scores object must fail validate\n%s", out)
	}
	mustContain(t, out, "alpha")
	mustContain(t, out, `"scores"`)
	if strings.Contains(out, "between 1 and 10") {
		t.Fatalf("a missing scores object must not be reported as a range fault\n%s", out)
	}
}

// Scenario: A wrongly typed scores value is reported as a shape error
func TestTV_Quality_WrongTypeScoresIsAShapeError(t *testing.T) {
	q := `{"role":"roadmap-quality-evaluator","threshold":9,` +
		`"features":[{"name":"alpha","scores":[1,2,3],"summary":"s"}]}`
	out, code := tvQualityFixture(t, q)
	if code == 0 {
		t.Fatalf("an array scores value must fail validate\n%s", out)
	}
	mustContain(t, out, "alpha")
	mustContain(t, out, "array")
	if strings.Contains(out, "between 1 and 10") {
		t.Fatalf("a shape fault must not be reported as a range fault\n%s", out)
	}
}

// Scenario: A missing individual score field names that field
func TestTV_Quality_MissingIndividualFieldNamesIt(t *testing.T) {
	partial := `{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,` +
		`"dependencies":9,"overall":9}`
	q := `{"role":"roadmap-quality-evaluator","threshold":9,` +
		`"features":[{"name":"alpha","scores":` + partial + `,"summary":"s"}]}`
	out, code := tvQualityFixture(t, q)
	if code == 0 {
		t.Fatalf("a missing individual score field must fail validate\n%s", out)
	}
	mustContain(t, out, "effortEstimation")
	if strings.Contains(out, "between 1 and 10") {
		t.Fatalf("a missing field must not be reported as a range fault\n%s", out)
	}
}

// Scenario: A non-integer score is reported as a type error naming the field
func TestTV_Quality_NonIntegerScoreNamesFieldAndType(t *testing.T) {
	bad := `{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,` +
		`"dependencies":9,"effortEstimation":9,"overall":9.0}`
	q := `{"role":"roadmap-quality-evaluator","threshold":9,` +
		`"features":[{"name":"alpha","scores":` + bad + `,"summary":"s"}]}`
	out, code := tvQualityFixture(t, q)
	if code == 0 {
		t.Fatalf("a non-integer score must fail validate\n%s", out)
	}
	mustContain(t, out, "overall")
	if strings.Contains(out, "between 1 and 10") {
		t.Fatalf("a type fault must not be reported as a range fault\n%s", out)
	}
}
