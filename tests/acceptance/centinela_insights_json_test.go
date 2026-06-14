package acceptance_test

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

// Acceptance: specs/centinela-insights.feature

// oneOfEach is a log with at least one event of every telemetry type.
var oneOfEach = []string{
	`{"type":"block","reason":"out-of-step","fileType":"plan","timestamp":"2026-01-01T00:00:00Z"}`,
	`{"type":"gate-failure","gate":"coverage","feature":"alpha","timestamp":"2026-02-01T00:00:00Z"}`,
	`{"type":"verify-rejection","feature":"alpha","timestamp":"2026-03-01T00:00:00Z"}`,
	`{"type":"complete-rejected","feature":"beta","timestamp":"2026-04-01T00:00:00Z"}`,
	`{"type":"step-advanced","feature":"alpha","timestamp":"2026-06-01T12:00:00Z"}`,
}

// Scenario: --json emits structured Report as indented JSON and exits 0
func TestInsightsJSONStructured(t *testing.T) {
	out, code := runInsights(t, insightsRepo(t, oneOfEach), "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	for _, f := range []string{"EventCount", "SpanStart", "SpanEnd", "Blocks", "Gates", "Rework", "StepsToGreen"} {
		if _, ok := m[f]; !ok {
			t.Fatalf("missing field %q: %v", f, m)
		}
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("JSON output contains ANSI: %q", out)
	}
}

// Scenario: --json output shape is stable across two runs on the same log
func TestInsightsJSONStableTwoRuns(t *testing.T) {
	dir := insightsRepo(t, oneOfEach)
	a, _ := runInsights(t, dir, "--json")
	b, _ := runInsights(t, dir, "--json")
	if a != b {
		t.Fatalf("json not byte-identical:\n%s\n---\n%s", a, b)
	}
}

// Scenario: --json output has stable field names usable by tooling
func TestInsightsJSONExactFields(t *testing.T) {
	out, _ := runInsights(t, insightsRepo(t, oneOfEach), "--json")
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	got := make([]string, 0, len(m))
	for k := range m {
		got = append(got, k)
	}
	sort.Strings(got)
	want := []string{"Blocks", "EventCount", "Gates", "Rework", "SpanEnd", "SpanStart", "StepsToGreen"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("top-level fields = %v, want %v", got, want)
	}
}
