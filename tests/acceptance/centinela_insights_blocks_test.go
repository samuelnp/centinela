package acceptance_test

import (
	"fmt"
	"strings"
	"testing"
)

// Acceptance: specs/centinela-insights.feature

func blockLine(reason, fileType string) string {
	return fmt.Sprintf(`{"type":"block","reason":%q,"fileType":%q}`, reason, fileType)
}

// indexOf returns the position of sub in s, or -1.
func indexOf(s, sub string) int { return strings.Index(s, sub) }

// Scenario: Blocks section ranks block events by count descending
func TestInsightsBlocksRanksByCountDesc(t *testing.T) {
	log := []string{
		blockLine("out-of-step", "plan"), blockLine("out-of-step", "plan"), blockLine("out-of-step", "plan"),
		blockLine("need-init", "source"), blockLine("need-init", "source"),
	}
	out, code := runInsights(t, insightsRepo(t, log))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "out-of-step · plan") || !strings.Contains(out, "need-init · source") {
		t.Fatalf("missing buckets: %q", out)
	}
	if indexOf(out, "out-of-step · plan") > indexOf(out, "need-init · source") {
		t.Fatalf("count 3 should precede count 2: %q", out)
	}
}

// Scenario: Blocks section respects --top N flag
func TestInsightsBlocksTopN(t *testing.T) {
	var log []string
	for i := 0; i < 10; i++ {
		log = append(log, blockLine(fmt.Sprintf("r%02d", i), "plan"))
	}
	out, _ := runInsights(t, insightsRepo(t, log), "--top", "3")
	if n := strings.Count(out, " · plan"); n != 3 {
		t.Fatalf("expected 3 block entries, got %d: %q", n, out)
	}
}

// Scenario: Default --top is 5 for the blocks section
func TestInsightsBlocksDefaultTopFive(t *testing.T) {
	var log []string
	for i := 0; i < 8; i++ {
		log = append(log, blockLine(fmt.Sprintf("r%02d", i), "plan"))
	}
	out, _ := runInsights(t, insightsRepo(t, log))
	if n := strings.Count(out, " · plan"); n != 5 {
		t.Fatalf("expected 5 block entries, got %d: %q", n, out)
	}
}

// Scenario: Blocks section ties break by key ascending for stable ordering
func TestInsightsBlocksTieBreakKeyAsc(t *testing.T) {
	out, _ := runInsights(t, insightsRepo(t, []string{blockLine("beta", "plan"), blockLine("alpha", "plan")}))
	if indexOf(out, "alpha · plan") > indexOf(out, "beta · plan") {
		t.Fatalf("alpha should precede beta: %q", out)
	}
}

// Scenario: Block event with empty fileType buckets under reason and empty-fileType key
func TestInsightsBlocksEmptyFileType(t *testing.T) {
	out, code := runInsights(t, insightsRepo(t, []string{`{"type":"block","reason":"out-of-step"}`}))
	if code != 0 || !strings.Contains(out, "out-of-step") {
		t.Fatalf("exit %d out %q", code, out)
	}
}

// Scenario: Log with only block events shows non-empty Blocks section and gracefully empty Gates and Rework sections
func TestInsightsOnlyBlocks(t *testing.T) {
	out, code := runInsights(t, insightsRepo(t, []string{blockLine("out-of-step", "plan")}))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	gates := sectionBody(out, "Gates")
	rework := sectionBody(out, "Rework")
	if !strings.Contains(gates, "(no events)") || !strings.Contains(rework, "(no events)") {
		t.Fatalf("Gates/Rework should be empty:\ngates=%q\nrework=%q", gates, rework)
	}
}

// Scenario: --top N larger than available block buckets returns all buckets without padding
func TestInsightsBlocksTopLargerThanBuckets(t *testing.T) {
	out, _ := runInsights(t, insightsRepo(t, []string{blockLine("a", "x"), blockLine("b", "x")}), "--top", "10")
	if n := strings.Count(out, " · x"); n != 2 {
		t.Fatalf("expected 2 entries, got %d: %q", n, out)
	}
}
