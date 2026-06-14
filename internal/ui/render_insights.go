package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/insights"
)

// RenderInsights renders the governance-telemetry analytics report in house
// style: a header (event count + span), then one section per metric. An empty
// log renders a single "no telemetry yet" line. Output is deterministic and
// lipgloss auto-strips ANSI on non-TTY, so piped output is plain and parseable.
func RenderInsights(r insights.Report) string {
	if r.EventCount == 0 {
		return StyleMuted.Render("no telemetry yet — run governed workflows to populate insights")
	}
	parts := []string{
		StyleBold.Render(fmt.Sprintf("Insights — %d events", r.EventCount)) +
			"  " + StyleMuted.Render(spanLabel(r.SpanStart, r.SpanEnd)),
		rankSection("Blocks (most-triggered)", r.Blocks),
		rankSection("Gates (most-failed)", r.Gates),
		rankSection("Rework (most friction)", r.Rework),
		stepsSection(r.StepsToGreen),
	}
	return strings.Join(parts, "\n\n")
}

// spanLabel renders the coverage window from the date portion of the RFC3339
// span timestamps, e.g. "2026-01-01 through 2026-06-01".
func spanLabel(start, end string) string {
	if start == "" || end == "" {
		return "span: (none)"
	}
	return "span: " + dateOnly(start) + " through " + dateOnly(end)
}

func dateOnly(ts string) string {
	if len(ts) >= 10 {
		return ts[:10]
	}
	return ts
}

// rankSection renders a titled, ranked Count list. An empty list renders
// "(no events)" so the section is never silently dropped.
func rankSection(title string, counts []insights.Count) string {
	lines := []string{StyleBold.Render(title)}
	if len(counts) == 0 {
		lines = append(lines, StyleMuted.Render("  (no events)"))
		return strings.Join(lines, "\n")
	}
	for _, c := range counts {
		lines = append(lines, fmt.Sprintf("  %s  %s",
			StyleMuted.Render(fmt.Sprintf("%3d", c.Count)), c.Key))
	}
	return strings.Join(lines, "\n")
}

// stepsSection renders the mean steps-to-green metric, "n/a" when undefined.
func stepsSection(s insights.StepsStat) string {
	mean := "n/a"
	if s.HasValue {
		mean = fmt.Sprintf("%.2f", s.Mean)
	}
	return StyleBold.Render("Steps-to-Green (mean attempts per advance)") + "\n" +
		fmt.Sprintf("  %s  %s",
			mean,
			StyleMuted.Render(fmt.Sprintf("(%d advances, %d rejections)", s.Advances, s.Rejections)))
}
