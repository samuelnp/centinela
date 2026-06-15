package acceptance_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// Acceptance: specs/capability-calibration.feature

// Scenario: --json emits structured Report as indented JSON and exits 0
func TestCalJSONStructured(t *testing.T) {
	lines := calConcat(calRepeat(3, func() string { return adv("claude-haiku-4-5") }),
		calRepeat(3, func() string { return gf("claude-haiku-4-5") }))
	out, code := runCal(t, calRepo(t, lines), "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	for _, f := range []string{"ModelCount", "SpanStart", "SpanEnd", "Models"} {
		if _, ok := m[f]; !ok {
			t.Fatalf("missing top-level field %q: %v", f, m)
		}
	}
	models, _ := m["Models"].([]any)
	if len(models) == 0 {
		t.Fatalf("expected a model entry: %v", m)
	}
	first, _ := models[0].(map[string]any)
	for _, f := range []string{"Model", "Class", "CurrentProfile", "Friction", "Recommendation", "RecommendedProfile", "Verdict"} {
		if _, ok := first[f]; !ok {
			t.Fatalf("model entry missing field %q: %v", f, first)
		}
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("JSON output contains ANSI: %q", out)
	}
}

// Scenario: --json on empty log emits a valid JSON Report with zero models and exits 0
func TestCalJSONEmptyLog(t *testing.T) {
	out, code := runCal(t, calRepo(t, nil), "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if c, _ := m["ModelCount"].(float64); c != 0 {
		t.Fatalf("ModelCount = %v, want 0", m["ModelCount"])
	}
	if models, ok := m["Models"].([]any); !ok || len(models) != 0 {
		t.Fatalf("Models should be empty array, got %v", m["Models"])
	}
}
