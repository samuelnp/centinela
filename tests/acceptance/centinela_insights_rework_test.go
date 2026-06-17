package acceptance_test

import (
	"fmt"
	"strings"
	"testing"
)

// Acceptance: specs/centinela-insights.feature

func featLine(typ, feature string) string {
	return fmt.Sprintf(`{"type":%q,"feature":%q}`, typ, feature)
}

// Scenario: Rework section ranks features by gate-failure plus verify-rejection plus complete-rejected count
func TestInsightsReworkRanksByCount(t *testing.T) {
	log := []string{
		featLine("gate-failure", "alpha"), featLine("gate-failure", "alpha"),
		featLine("verify-rejection", "alpha"),
		featLine("complete-rejected", "beta"), featLine("gate-failure", "beta"),
	}
	out, code := runInsights(t, insightsRepo(t, log))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	r := sectionBody(out, "Rework")
	if !strings.Contains(r, "alpha") || !strings.Contains(r, "beta") {
		t.Fatalf("missing features: %q", r)
	}
	if indexOf(r, "alpha") > indexOf(r, "beta") {
		t.Fatalf("alpha (3) should precede beta (2): %q", r)
	}
}

// Scenario: Rework section excludes events with no feature field
func TestInsightsReworkExcludesEmptyFeature(t *testing.T) {
	log := []string{featLine("gate-failure", ""), featLine("gate-failure", "alpha")}
	out, _ := runInsights(t, insightsRepo(t, log))
	r := sectionBody(out, "Rework")
	if !strings.Contains(r, "alpha") {
		t.Fatalf("expected alpha: %q", r)
	}
	// exactly one entry line under the title (no empty-feature bucket).
	if n := strings.Count(r, "\n  "); n != 1 {
		t.Fatalf("expected exactly 1 rework entry, got %d: %q", n, r)
	}
}

// Scenario: Rework section respects --top N flag
func TestInsightsReworkTopN(t *testing.T) {
	var log []string
	for i := 0; i < 6; i++ {
		log = append(log, featLine("gate-failure", fmt.Sprintf("f%02d", i)))
	}
	out, _ := runInsights(t, insightsRepo(t, log), "--top", "3")
	r := sectionBody(out, "Rework")
	if n := strings.Count(r, "  f"); n != 3 {
		t.Fatalf("expected 3 rework entries, got %d: %q", n, r)
	}
}

// Scenario: Rework section ties break by feature name ascending for stable ordering
func TestInsightsReworkTieBreakFeatureAsc(t *testing.T) {
	out, _ := runInsights(t, insightsRepo(t, []string{featLine("gate-failure", "zeta"), featLine("gate-failure", "alpha")}))
	r := sectionBody(out, "Rework")
	if indexOf(r, "alpha") > indexOf(r, "zeta") {
		t.Fatalf("alpha should precede zeta: %q", r)
	}
}

// Scenario: step-advanced events are not counted in rework score
func TestInsightsReworkExcludesStepAdvanced(t *testing.T) {
	log := []string{
		featLine("step-advanced", "alpha"), featLine("step-advanced", "alpha"),
		featLine("gate-failure", "alpha"),
	}
	out, _ := runInsights(t, insightsRepo(t, log))
	r := sectionBody(out, "Rework")
	// alpha rework count must be 1, not 3 (step-advanced excluded).
	if !strings.Contains(r, "1  alpha") {
		t.Fatalf("expected 'alpha' count 1: %q", r)
	}
}
