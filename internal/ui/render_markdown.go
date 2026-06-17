package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/gates"
)

// MarkdownMarker is the stable first line the CI poster greps to find and
// update its single PR comment.
const MarkdownMarker = "<!-- centinela:pr-gate -->"

const detailsCap = 50

// RenderGatesMarkdown renders gate results as deterministic, plain Markdown
// (no Lipgloss/ANSI). The marker line is first, followed by a header with
// pass/fail/warn counts, a table of gates, and a <details> block per failing
// gate (Details capped). Results are emitted in input order.
func RenderGatesMarkdown(results []gates.Result) string {
	var b strings.Builder
	b.WriteString(MarkdownMarker + "\n")
	b.WriteString(markdownHeader(results) + "\n\n")
	b.WriteString("| Gate | Status | Message |\n| --- | --- | --- |\n")
	for _, r := range results {
		icon, label := mdStatus(r.Status)
		b.WriteString(fmt.Sprintf("| %s | %s %s | %s |\n", r.Name, icon, label, r.Message))
	}
	for _, r := range results {
		if r.Status == gates.Fail {
			b.WriteString("\n" + mdDetails(r))
		}
	}
	return b.String()
}

func markdownHeader(results []gates.Result) string {
	var pass, fail, warn int
	for _, r := range results {
		switch r.Status {
		case gates.Pass:
			pass++
		case gates.Fail:
			fail++
		case gates.Warn:
			warn++
		}
	}
	icon := "✅"
	if fail > 0 {
		icon = "❌"
	} else if warn > 0 {
		icon = "⚠️"
	}
	return fmt.Sprintf("### Centinela PR Gate — %s %d failed, %d passed, %d warned",
		icon, fail, pass, warn)
}

func mdDetails(r gates.Result) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<details><summary>Failing details (%s)</summary>\n\n", r.Name))
	shown := r.Details
	extra := 0
	if len(shown) > detailsCap {
		extra = len(shown) - detailsCap
		shown = shown[:detailsCap]
	}
	for _, d := range shown {
		b.WriteString("- " + d + "\n")
	}
	if extra > 0 {
		b.WriteString(fmt.Sprintf("- … %d more\n", extra))
	}
	b.WriteString("</details>\n")
	return b.String()
}

// mdStatus maps a gate status to its (icon, label) pair for the table cell.
func mdStatus(s gates.Status) (icon, label string) {
	switch s {
	case gates.Pass:
		return "✅", "pass"
	case gates.Fail:
		return "❌", "fail"
	case gates.Warn:
		return "⚠️", "warn"
	default:
		return "➖", "skip"
	}
}
