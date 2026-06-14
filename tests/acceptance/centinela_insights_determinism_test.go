package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/centinela-insights.feature

// Scenario: Two runs on the same log produce byte-identical human output
func TestInsightsHumanByteIdentical(t *testing.T) {
	dir := insightsRepo(t, oneOfEach)
	a, _ := runInsights(t, dir)
	b, _ := runInsights(t, dir)
	if a != b {
		t.Fatalf("human output not byte-identical:\n%s\n---\n%s", a, b)
	}
}

// Scenario: Ties between buckets with equal count are always broken by key ascending
func TestInsightsTiesBrokenByKeyAsc(t *testing.T) {
	log := []string{
		gateLine("z-gate"), gateLine("z-gate"),
		gateLine("a-gate"), gateLine("a-gate"),
		gateLine("m-gate"), gateLine("m-gate"),
	}
	out, _ := runInsights(t, insightsRepo(t, log))
	g := sectionBody(out, "Gates")
	ia, im, iz := indexOf(g, "a-gate"), indexOf(g, "m-gate"), indexOf(g, "z-gate")
	if !(ia < im && im < iz) {
		t.Fatalf("expected a-gate, m-gate, z-gate order: %q", g)
	}
}

// Scenario: Malformed JSONL lines are skipped and valid events are still aggregated
func TestInsightsMalformedLinesSkipped(t *testing.T) {
	log := []string{
		blockLine("out-of-step", "plan"),
		`{ this is not valid json`,
		gateLine("coverage"),
	}
	out, code := runInsights(t, insightsRepo(t, log))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(sectionBody(out, "Blocks"), "out-of-step") {
		t.Fatalf("valid block lost: %q", out)
	}
	if !strings.Contains(sectionBody(out, "Gates"), "coverage") {
		t.Fatalf("valid gate lost: %q", out)
	}
}

// Scenario: A log with a single event of each type is reported without crash
func TestInsightsSingleOfEachType(t *testing.T) {
	out, code := runInsights(t, insightsRepo(t, oneOfEach))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if strings.Contains(sectionBody(out, "Blocks"), "(no events)") {
		t.Fatalf("Blocks should have one entry: %q", out)
	}
	if strings.Contains(sectionBody(out, "Gates"), "(no events)") {
		t.Fatalf("Gates should have one entry: %q", out)
	}
	if !strings.Contains(sectionBody(out, "Rework"), "alpha") {
		t.Fatalf("Rework should list alpha: %q", out)
	}
	// 1 advance, 1 rejection ⇒ 2.00.
	if !strings.Contains(stepsBody(out), "2.00") {
		t.Fatalf("expected steps 2.00: %q", out)
	}
}
