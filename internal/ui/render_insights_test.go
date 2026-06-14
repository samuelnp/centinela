package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/insights"
)

// An empty-state report renders the "no telemetry yet" line and nothing else.
func TestRenderInsightsEmptyState(t *testing.T) {
	out := RenderInsights(insights.Report{})
	if !strings.Contains(out, "no telemetry yet") {
		t.Fatalf("missing empty-state line: %q", out)
	}
	if strings.Contains(out, "Blocks") {
		t.Fatalf("empty state should not render sections: %q", out)
	}
}

// A populated report renders the header, span, all four sections, and counts.
func TestRenderInsightsSections(t *testing.T) {
	r := insights.Report{
		EventCount: 6, SpanStart: "2026-01-01T00:00:00Z", SpanEnd: "2026-06-01T12:00:00Z",
		Blocks:       []insights.Count{{Key: "out-of-step · plan", Count: 2}},
		Gates:        []insights.Count{{Key: "coverage", Count: 1}},
		Rework:       []insights.Count{{Key: "alpha", Count: 2}},
		StepsToGreen: insights.StepsStat{Advances: 1, Rejections: 1, Mean: 2.0, HasValue: true},
	}
	out := RenderInsights(r)
	for _, want := range []string{
		"Insights — 6 events", "2026-01-01 through 2026-06-01",
		"Blocks", "Gates", "Rework", "Steps-to-Green",
		"out-of-step · plan", "coverage", "alpha", "2.00",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
}

// A section with no entries renders "(no events)" rather than disappearing.
func TestRenderInsightsEmptySection(t *testing.T) {
	r := insights.Report{
		EventCount: 1, SpanStart: "2026-01-01T00:00:00Z", SpanEnd: "2026-01-01T00:00:00Z",
		Blocks:       []insights.Count{{Key: "x", Count: 1}},
		StepsToGreen: insights.StepsStat{},
	}
	out := RenderInsights(r)
	if !strings.Contains(out, "(no events)") {
		t.Fatalf("empty Gates/Rework should render (no events): %q", out)
	}
}

// Undefined steps-to-green (zero advances) renders "n/a".
func TestRenderInsightsStepsNA(t *testing.T) {
	r := insights.Report{
		EventCount: 1, SpanStart: "2026-01-01T00:00:00Z", SpanEnd: "2026-01-01T00:00:00Z",
		StepsToGreen: insights.StepsStat{Rejections: 1},
	}
	if out := RenderInsights(r); !strings.Contains(out, "n/a") {
		t.Fatalf("expected n/a steps: %q", out)
	}
}

// A report with no span timestamps renders "(none)".
func TestRenderInsightsNoSpan(t *testing.T) {
	r := insights.Report{EventCount: 1, StepsToGreen: insights.StepsStat{Advances: 1, Mean: 1, HasValue: true}}
	if out := RenderInsights(r); !strings.Contains(out, "(none)") {
		t.Fatalf("expected (none) span: %q", out)
	}
}

// Rendered output never contains raw ANSI escape sequences in tests (non-TTY).
func TestRenderInsightsNoANSI(t *testing.T) {
	r := insights.Report{EventCount: 1, Blocks: []insights.Count{{Key: "x", Count: 1}}}
	if strings.Contains(RenderInsights(r), "\x1b[") {
		t.Fatal("output contains ANSI escape sequences")
	}
}
