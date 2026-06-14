package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/insights"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/ui"
)

// writeLog writes a multi-type events.jsonl into dir/.workflow/telemetry.
func writeLog(t *testing.T, dir string) {
	t.Helper()
	td := filepath.Join(dir, ".workflow", "telemetry")
	if err := os.MkdirAll(td, 0o755); err != nil {
		t.Fatal(err)
	}
	lines := []string{
		`{"type":"block","reason":"out-of-step","fileType":"plan","timestamp":"2026-01-01T00:00:00Z"}`,
		`{"type":"block","reason":"out-of-step","fileType":"plan","timestamp":"2026-01-02T00:00:00Z"}`,
		`{"type":"gate-failure","gate":"coverage","feature":"alpha","timestamp":"2026-02-01T00:00:00Z"}`,
		`{"type":"verify-rejection","feature":"alpha","timestamp":"2026-03-01T00:00:00Z"}`,
		`{"type":"complete-rejected","feature":"beta","timestamp":"2026-04-01T00:00:00Z"}`,
		`{"type":"step-advanced","feature":"alpha","timestamp":"2026-06-01T12:00:00Z"}`,
		`garbage not json`,
	}
	if err := os.WriteFile(filepath.Join(td, "events.jsonl"), []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Full pipeline: Read the JSONL log, Compute, render human + JSON; the garbage
// line is skipped and all four metrics aggregate correctly.
func TestInsightsPipelineAndJSONRoundTrip(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir)

	events, err := telemetry.Read(filepath.Join(dir, ".workflow", "telemetry"))
	if err != nil {
		t.Fatalf("telemetry.Read: %v", err)
	}
	if len(events) != 6 {
		t.Fatalf("expected 6 valid events (garbage skipped), got %d", len(events))
	}

	r := insights.Compute(events, 5)
	if r.EventCount != 6 || r.SpanStart != "2026-01-01T00:00:00Z" || r.SpanEnd != "2026-06-01T12:00:00Z" {
		t.Fatalf("report header wrong: %+v", r)
	}
	if len(r.Blocks) != 1 || r.Blocks[0].Count != 2 {
		t.Fatalf("blocks = %+v", r.Blocks)
	}
	if len(r.Rework) != 2 || r.Rework[0].Key != "alpha" || r.Rework[0].Count != 2 {
		t.Fatalf("rework = %+v", r.Rework)
	}
	if !r.StepsToGreen.HasValue || r.StepsToGreen.Mean != 2.0 {
		t.Fatalf("steps = %+v", r.StepsToGreen)
	}

	human := ui.RenderInsights(r)
	if !strings.Contains(human, "Insights — 6 events") || strings.Contains(human, "\x1b[") {
		t.Fatalf("human render wrong or has ANSI: %q", human)
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back insights.Report
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.EventCount != r.EventCount || back.Rework[0].Key != "alpha" {
		t.Fatalf("json round-trip lost data: %+v", back)
	}
}

// Missing log ⇒ empty events, empty-state report, no error.
func TestInsightsPipelineEmptyLog(t *testing.T) {
	events, err := telemetry.Read(filepath.Join(t.TempDir(), "nope"))
	if err != nil {
		t.Fatalf("Read missing: %v", err)
	}
	r := insights.Compute(events, 5)
	if r.EventCount != 0 {
		t.Fatalf("expected 0 events, got %d", r.EventCount)
	}
	if !strings.Contains(ui.RenderInsights(r), "no telemetry yet") {
		t.Fatal("expected empty-state line")
	}
}
