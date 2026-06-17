package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/centinela-insights.feature

// Scenario: Piped output contains no ANSI escape sequences
func TestInsightsPipedNoANSI(t *testing.T) {
	// runCent captures via CombinedOutput (a pipe, not a TTY), which is exactly
	// the `centinela insights | cat` condition: lipgloss must strip styling.
	out, code := runInsights(t, insightsRepo(t, oneOfEach))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("output contains ANSI escape sequences: %q", out)
	}
	// Plain text: every rune is printable/whitespace so grep/awk can parse it.
	for _, r := range out {
		if r < 0x20 && r != '\n' && r != '\t' && r != '\r' {
			t.Fatalf("non-printable control rune %#U in output", r)
		}
	}
}

// Scenario: Report includes span of earliest and latest event timestamps
func TestInsightsSpanRange(t *testing.T) {
	log := []string{
		`{"type":"block","reason":"r","fileType":"f","timestamp":"2026-06-01T12:00:00Z"}`,
		`{"type":"block","reason":"r","fileType":"f","timestamp":"2026-01-01T00:00:00Z"}`,
	}
	out, _ := runInsights(t, insightsRepo(t, log))
	if !strings.Contains(out, "2026-01-01") || !strings.Contains(out, "2026-06-01") {
		t.Fatalf("expected span 2026-01-01 through 2026-06-01: %q", out)
	}
	if indexOf(out, "2026-01-01") > indexOf(out, "2026-06-01") {
		t.Fatalf("start should precede end in span line: %q", out)
	}
}

// Scenario: Report includes total event count considered
func TestInsightsTotalEventCount(t *testing.T) {
	log := []string{
		gateLine("a"), gateLine("b"), gateLine("c"), gateLine("d"),
		blockLine("r", "f"), blockLine("r", "g"), `{"type":"step-advanced"}`,
	}
	out, code := runInsights(t, insightsRepo(t, log))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "7 events") {
		t.Fatalf("expected total count 7: %q", out)
	}
}
