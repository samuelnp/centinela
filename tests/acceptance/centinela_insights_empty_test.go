package acceptance_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// Acceptance: specs/centinela-insights.feature

// Scenario: Missing telemetry log prints clean empty-state report and exits 0
func TestInsightsMissingLogEmptyState(t *testing.T) {
	out, code := runInsights(t, insightsRepo(t, nil))
	if code != 0 {
		t.Fatalf("exit %d, want 0: %s", code, out)
	}
	if !strings.Contains(out, "no telemetry yet") {
		t.Fatalf("missing empty-state line: %q", out)
	}
	for _, bad := range []string{"panic", "goroutine", "Error:"} {
		if strings.Contains(out, bad) {
			t.Fatalf("output contains %q: %q", bad, out)
		}
	}
}

// Scenario: Empty telemetry log prints clean empty-state report and exits 0
func TestInsightsEmptyLogEmptyState(t *testing.T) {
	out, code := runInsights(t, insightsRepo(t, []string{}))
	if code != 0 || !strings.Contains(out, "no telemetry yet") {
		t.Fatalf("exit %d out %q", code, out)
	}
}

// Scenario: Whitespace-only telemetry log is treated as empty and exits 0
func TestInsightsWhitespaceLogEmptyState(t *testing.T) {
	out, code := runInsights(t, insightsRepo(t, []string{"", "   ", "\t"}))
	if code != 0 || !strings.Contains(out, "no telemetry yet") {
		t.Fatalf("exit %d out %q", code, out)
	}
}

// Scenario: --json on empty log emits a valid JSON Report with zero counts and exits 0
func TestInsightsJSONEmptyLogZeroCount(t *testing.T) {
	out, code := runInsights(t, insightsRepo(t, nil), "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if c, _ := m["EventCount"].(float64); c != 0 {
		t.Fatalf("EventCount = %v, want 0", m["EventCount"])
	}
}
