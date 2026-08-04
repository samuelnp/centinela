// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const gbdBacklogRoadmap = `{"phases":[{"name":"Phase 1: Foundations","features":[]},` +
	`{"name":"Backlog","features":[{"name":"low-score-finding","summary":"needs promotion",` +
	`"source":{"feature":"guided-by-default","role":"qa-senior"},"deferredAt":"2026-01-01T00:00:00Z"}]}]}`

// Scenario: Promoting a feature with a low overall score succeeds
func TestGBD_PromoteWithLowScoreSucceeds(t *testing.T) {
	bin := buildCent(t)
	dir := acceptanceDir(t, gbdBacklogRoadmap)
	seedPromoteArtifacts(t, dir)
	out, code := runCent(t, bin, dir, "roadmap", "promote", "low-score-finding",
		"--phase", "Phase 1: Foundations", "--scores", "3,3,3,3,3,3")
	if code != 0 {
		t.Fatalf("promote with overall 3 must succeed, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Promoted") {
		t.Fatalf("expected a promotion success message: %s", out)
	}
}

// Scenario: Out-of-range scores are still refused
func TestGBD_OutOfRangeScoreRefused(t *testing.T) {
	for _, score := range []int{0, 11} {
		scores := fmt.Sprintf(`{"acceptanceCriteria":9,"userValue":%d,"definitionClarity":9,`+
			`"dependencies":9,"effortEstimation":9,"overall":9}`, score)
		q := `{"role":"roadmap-quality-evaluator","threshold":9,"features":[{"name":"alpha","scores":` +
			scores + `,"summary":"s"}]}`
		out, code := tvQualityFixture(t, q)
		if code == 0 {
			t.Fatalf("score %d must be refused: %s", score, out)
		}
		if !strings.Contains(out, "between 1 and 10") || !strings.Contains(out, "userValue") {
			t.Fatalf("refusal must name the offending field and range: %s", out)
		}
	}
}

// Scenario: Existing quality artifacts survive the deletion untouched
func TestGBD_ExistingQualityArtifactNotRewritten(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	writeFile(t, dir, ".workflow/roadmap.json", tvRoadmapJSON)
	writeFile(t, dir, ".workflow/roadmap-analysis.json", tvAnalysisJSON)
	writeFile(t, dir, ".workflow/roadmap-analysis.md", "# a\n")
	writeFile(t, dir, ".workflow/roadmap-quality.md", "# q\n")
	q := `{"role":"roadmap-quality-evaluator","threshold":9,"features":[{"name":"alpha","scores":` +
		tvFullScores + `,"summary":"s"}]}`
	writeFile(t, dir, ".workflow/roadmap-quality.json", q)
	before, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap-quality.json"))

	out, code := runCent(t, bin, dir, "roadmap", "validate")
	if code != 0 {
		t.Fatalf("validate must pass: %s", out)
	}
	after, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap-quality.json"))
	if string(before) != string(after) {
		t.Fatal("roadmap validate must not rewrite an existing quality report")
	}
}
