package acceptance_test

import (
	"fmt"
	"strings"
	"testing"
)

// Acceptance: specs/centinela-insights.feature

func gateLine(gate string) string { return fmt.Sprintf(`{"type":"gate-failure","gate":%q}`, gate) }

// Scenario: Gates section ranks gate-failure events by count descending
func TestInsightsGatesRanksByCountDesc(t *testing.T) {
	log := []string{gateLine("coverage"), gateLine("coverage"), gateLine("import-graph")}
	out, code := runInsights(t, insightsRepo(t, log))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	g := sectionBody(out, "Gates")
	if !strings.Contains(g, "coverage") || !strings.Contains(g, "import-graph") {
		t.Fatalf("missing gates: %q", g)
	}
	if indexOf(g, "coverage") > indexOf(g, "import-graph") {
		t.Fatalf("coverage (2) should precede import-graph (1): %q", g)
	}
}

// Scenario: Gates section respects --top N flag
func TestInsightsGatesTopN(t *testing.T) {
	var log []string
	for i := 0; i < 7; i++ {
		log = append(log, gateLine(fmt.Sprintf("g%02d", i)))
	}
	out, _ := runInsights(t, insightsRepo(t, log), "--top", "2")
	g := sectionBody(out, "Gates")
	// 2 gate lines, each "    1  g.." — count the indented count tokens.
	if n := strings.Count(g, "  g"); n != 2 {
		t.Fatalf("expected 2 gate entries, got %d: %q", n, g)
	}
}

// Scenario: Gate-failure event with empty Gate field buckets under key rendered as none
func TestInsightsGatesEmptyGateNone(t *testing.T) {
	out, code := runInsights(t, insightsRepo(t, []string{`{"type":"gate-failure","gate":""}`}))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(sectionBody(out, "Gates"), "<none>") {
		t.Fatalf("expected <none> gate: %q", out)
	}
}

// Scenario: Log with only gate-failure events shows non-empty Gates section and gracefully empty other sections
func TestInsightsOnlyGates(t *testing.T) {
	out, code := runInsights(t, insightsRepo(t, []string{gateLine("coverage")}))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if strings.Contains(sectionBody(out, "Gates"), "(no events)") {
		t.Fatalf("Gates should be non-empty: %q", out)
	}
	if !strings.Contains(sectionBody(out, "Blocks"), "(no events)") ||
		!strings.Contains(sectionBody(out, "Rework"), "(no events)") {
		t.Fatalf("Blocks/Rework should be empty: %q", out)
	}
}

// Scenario: Gates section ties break by gate name ascending for stable ordering
func TestInsightsGatesTieBreakKeyAsc(t *testing.T) {
	out, _ := runInsights(t, insightsRepo(t, []string{gateLine("security"), gateLine("coverage")}))
	g := sectionBody(out, "Gates")
	if indexOf(g, "coverage") > indexOf(g, "security") {
		t.Fatalf("coverage should precede security: %q", g)
	}
}
