// Acceptance: specs/guided-by-default.feature
//
// Reuses tvQualityFixture / tvRoadmapJSON / tvAnalysisJSON / tvFullScores
// from truthful_validators_quality_test.go — the same fixture shape, now
// asserted under THIS spec's scenario names so guided-by-default's own
// traceability closes without a second harness.
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: A low self-assigned quality score no longer blocks anything
func TestGBD_LowQualityScoreDoesNotBlock(t *testing.T) {
	q := `{"role":"roadmap-quality-evaluator","threshold":9,"features":[{"name":"alpha","scores":` +
		`{"acceptanceCriteria":3,"userValue":3,"definitionClarity":3,"dependencies":3,"effortEstimation":3,"overall":3},"summary":"s"}]}`
	out, code := tvQualityFixture(t, q)
	if code != 0 {
		t.Fatalf("overall:3 must still validate, got %d: %s", code, out)
	}
	if !strings.Contains(out, "valid") {
		t.Fatalf("expected the roadmap reported as valid: %s", out)
	}
	if !strings.Contains(out, "Advisory") {
		t.Fatalf("the low score must be surfaced as advice, not silence: %s", out)
	}
}

// Scenario: The declared threshold field is no longer enforced
func TestGBD_DeclaredThresholdFieldIgnored(t *testing.T) {
	q := `{"role":"roadmap-quality-evaluator","threshold":3,"features":[{"name":"alpha","scores":` +
		tvFullScores + `,"summary":"s"}]}`
	out, code := tvQualityFixture(t, q)
	if code != 0 {
		t.Fatalf("a non-9 declared threshold must not fail validate, got %d: %s", code, out)
	}
	if !strings.Contains(out, "valid") {
		t.Fatalf("expected the roadmap reported as valid: %s", out)
	}
}

// Scenario: A missing scores object is still reported as a shape fault
func TestGBD_MissingScoresObjectIsShapeFault(t *testing.T) {
	q := `{"role":"roadmap-quality-evaluator","threshold":9,"features":[{"name":"alpha","summary":"s"}]}`
	out, code := tvQualityFixture(t, q)
	if code == 0 {
		t.Fatalf("a missing scores object must fail validate: %s", out)
	}
	if !strings.Contains(out, "scores") {
		t.Fatalf("refusal must name the missing scores field: %s", out)
	}
	if strings.Contains(out, "between 1 and 10") {
		t.Fatalf("must not be reported as a range fault: %s", out)
	}
}

// Scenario: A scores field of the wrong JSON kind is still reported as a shape fault
func TestGBD_WrongKindScoresIsShapeFault(t *testing.T) {
	q := `{"role":"roadmap-quality-evaluator","threshold":9,"features":[{"name":"alpha","scores":[1,2,3],"summary":"s"}]}`
	out, code := tvQualityFixture(t, q)
	if code == 0 {
		t.Fatalf("an array scores value must fail validate: %s", out)
	}
	if !strings.Contains(out, "array") {
		t.Fatalf("refusal must name the JSON kind found: %s", out)
	}
	if strings.Contains(out, "between 1 and 10") {
		t.Fatalf("a shape fault must not be reported as a range fault: %s", out)
	}
}

// Scenario: A non-integer score is still reported as a type fault
func TestGBD_NonIntegerScoreIsTypeFault(t *testing.T) {
	bad := `{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,` +
		`"dependencies":9,"effortEstimation":9,"overall":"nine"}`
	q := `{"role":"roadmap-quality-evaluator","threshold":9,"features":[{"name":"alpha","scores":` +
		bad + `,"summary":"s"}]}`
	out, code := tvQualityFixture(t, q)
	if code == 0 {
		t.Fatalf("a non-integer overall must fail validate: %s", out)
	}
	if !strings.Contains(out, "overall") {
		t.Fatalf("refusal must name the offending field: %s", out)
	}
}
