package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestRunCalibrateMissingLogExitsZero — no log → empty-state, no error (exit 0).
func TestRunCalibrateMissingLogExitsZero(t *testing.T) {
	out, err := captureCalibrate(t, t.TempDir(), false)
	if err != nil {
		t.Fatalf("missing log should not error: %v", err)
	}
	if !strings.Contains(out, "no telemetry yet") {
		t.Fatalf("missing empty-state line: %q", out)
	}
}

// TestRunCalibrateHumanExitsZero — populated log renders a model block, no error.
func TestRunCalibrateHumanExitsZero(t *testing.T) {
	dir := t.TempDir()
	seedCalLog(t, dir,
		`{"type":"step-advanced","model":"claude-haiku-4-5","timestamp":"2026-01-01T00:00:00Z"}`,
		`{"type":"step-advanced","model":"claude-haiku-4-5","timestamp":"2026-01-02T00:00:00Z"}`,
		`{"type":"step-advanced","model":"claude-haiku-4-5","timestamp":"2026-01-03T00:00:00Z"}`,
		`{"type":"gate-failure","model":"claude-haiku-4-5","timestamp":"2026-01-04T00:00:00Z"}`,
	)
	out, err := captureCalibrate(t, dir, false)
	if err != nil || !strings.Contains(out, "claude-haiku-4-5") {
		t.Fatalf("human render wrong: err=%v out=%q", err, out)
	}
}

// TestRunCalibrateJSONValidAndStable — --json emits valid, byte-stable JSON with
// the documented top-level fields and no ANSI.
func TestRunCalibrateJSONValidAndStable(t *testing.T) {
	dir := t.TempDir()
	seedCalLog(t, dir,
		`{"type":"step-advanced","model":"claude-haiku-4-5","timestamp":"2026-01-01T00:00:00Z"}`,
		`{"type":"gate-failure","model":"claude-haiku-4-5","timestamp":"2026-01-02T00:00:00Z"}`,
	)
	out, err := captureCalibrate(t, dir, true)
	if err != nil {
		t.Fatalf("json run errored: %v", err)
	}
	var m map[string]any
	if e := json.Unmarshal([]byte(out), &m); e != nil {
		t.Fatalf("invalid JSON: %v\n%s", e, out)
	}
	for _, f := range []string{"ModelCount", "SpanStart", "SpanEnd", "Models"} {
		if _, ok := m[f]; !ok {
			t.Fatalf("missing field %q", f)
		}
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("json contains ANSI: %q", out)
	}
	out2, _ := captureCalibrate(t, dir, true)
	if out != out2 {
		t.Fatalf("json not stable:\n%s\n---\n%s", out, out2)
	}
}
