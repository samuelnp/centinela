package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runInsightsCapture chdirs into dir, runs the insights command, and returns
// everything written to stdout. Restores cwd and flags afterwards.
func runInsightsCapture(t *testing.T, dir string, top int, asJSON bool) string {
	t.Helper()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	insightsTop, insightsJSON = top, asJSON
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	err := runInsights(nil, nil)
	_ = w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("runInsights: %v", err)
	}
	buf := make([]byte, 64*1024)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

// seedLog writes events.jsonl under dir/.workflow/telemetry.
func seedLog(t *testing.T, dir, content string) {
	t.Helper()
	td := filepath.Join(dir, ".workflow", "telemetry")
	if err := os.MkdirAll(td, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(td, "events.jsonl"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Missing log ⇒ empty-state report, no error (exit 0).
func TestInsightsMissingLog(t *testing.T) {
	if out := runInsightsCapture(t, t.TempDir(), 5, false); !strings.Contains(out, "no telemetry yet") {
		t.Fatalf("expected empty-state: %q", out)
	}
}

// --top truncates ranked sections.
func TestInsightsTopTruncates(t *testing.T) {
	dir := t.TempDir()
	var b strings.Builder
	for _, r := range []string{"a", "b", "c", "d", "e"} {
		b.WriteString(`{"type":"block","reason":"` + r + `"}` + "\n")
	}
	seedLog(t, dir, b.String())
	out := runInsightsCapture(t, dir, 2, false)
	// 5 distinct buckets, top 2 ⇒ only 2 block lines (each " · <none>").
	if n := strings.Count(out, " · <none>"); n != 2 {
		t.Fatalf("expected 2 block entries, got %d: %q", n, out)
	}
}

// --json emits valid, stable JSON with the Report fields.
func TestInsightsJSONValidAndStable(t *testing.T) {
	dir := t.TempDir()
	seedLog(t, dir, `{"type":"gate-failure","gate":"coverage","feature":"x","timestamp":"2026-01-01T00:00:00Z"}`+"\n")
	a := runInsightsCapture(t, dir, 5, true)
	b := runInsightsCapture(t, dir, 5, true)
	if a != b {
		t.Fatalf("json not stable:\n%s\n---\n%s", a, b)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(a), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	for _, f := range []string{"EventCount", "SpanStart", "SpanEnd", "Blocks", "Gates", "Rework", "StepsToGreen"} {
		if _, ok := m[f]; !ok {
			t.Fatalf("missing field %q in %v", f, m)
		}
	}
}
