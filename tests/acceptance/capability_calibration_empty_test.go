package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/capability-calibration.feature

// Scenario: Missing telemetry log prints clean empty-state report and exits 0
func TestCalMissingLog(t *testing.T) {
	out, code := runCal(t, calRepo(t, nil))
	if code != 0 || !strings.Contains(out, "no telemetry yet") {
		t.Fatalf("missing-log wrong (code %d):\n%s", code, out)
	}
	for _, bad := range []string{"panic", "goroutine", "Error:"} {
		if strings.Contains(out, bad) {
			t.Fatalf("output contains %q:\n%s", bad, out)
		}
	}
}

// Scenario: Empty telemetry log prints clean empty-state report and exits 0
func TestCalEmptyLog(t *testing.T) {
	out, code := runCal(t, calRepo(t, []string{}))
	if code != 0 || !strings.Contains(out, "no telemetry yet") {
		t.Fatalf("empty-log wrong (code %d):\n%s", code, out)
	}
}

// Scenario: Malformed JSONL lines are skipped and valid events are still aggregated
func TestCalMalformedSkipped(t *testing.T) {
	lines := []string{
		adv("claude-sonnet-4-6"),
		`{not valid json`,
		adv("claude-sonnet-4-6"),
	}
	out, code := runCal(t, calRepo(t, lines))
	sec := recordSection(out, "claude-sonnet-4-6")
	if code != 0 || !strings.Contains(sec, "advances=2") {
		t.Fatalf("malformed-skip wrong (code %d):\n%s", code, sec)
	}
	if strings.Contains(out, "parse error") || strings.Contains(out, "invalid character") {
		t.Fatalf("parse error leaked:\n%s", out)
	}
}
