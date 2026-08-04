package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// summaryWidth keeps the table terminal-friendly on an 80-column terminal
// after the age, slug and source columns.
const summaryWidth = 44

// RenderBacklogList renders the aged Backlog as an oldest-first table plus the
// footer arithmetic. filtered says the rows were narrowed to stale findings, so
// the empty state can name the threshold instead of claiming an empty Backlog.
func RenderBacklogList(rows []roadmap.Aged, stats roadmap.BacklogStats, filtered bool) string {
	if len(rows) == 0 {
		return StyleMuted.Render(emptyBacklogLine(stats, filtered))
	}
	lines := make([]string, 0, len(rows)+2)
	for _, a := range rows {
		lines = append(lines, fmt.Sprintf("  %-8s %-38s %s", ageLabel(a),
			a.Name, StyleMuted.Render(backlogAnnotation(a))))
	}
	lines = append(lines, "", StyleMuted.Render(renderBacklogFooter(stats)))
	return strings.Join(lines, "\n")
}

// emptyBacklogLine distinguishes "nothing deferred at all" from "nothing has
// rotted yet" — the second is good news about a non-empty Backlog.
func emptyBacklogLine(stats roadmap.BacklogStats, filtered bool) string {
	if filtered && stats.Total > 0 {
		return fmt.Sprintf("No findings older than %dd (%d in the Backlog).",
			stats.ThresholdDays, stats.Total)
	}
	return "No deferred findings in the Backlog."
}

// renderBacklogFooter states the totals and names the oldest finding.
func renderBacklogFooter(stats roadmap.BacklogStats) string {
	footer := fmt.Sprintf("%d findings · %d older than %dd",
		stats.Total, stats.Stale, stats.ThresholdDays)
	if stats.OldestKnown {
		footer += fmt.Sprintf(" · oldest %dd (%s)", stats.OldestDays, stats.OldestSlug)
	}
	return footer
}

// ageLabel renders the age column, "unknown" when deferredAt was unreadable.
func ageLabel(a roadmap.Aged) string {
	if !a.KnownAge {
		return "unknown"
	}
	return fmt.Sprintf("%dd", a.AgeDays)
}

// backlogAnnotation renders the source provenance and the truncated summary.
func backlogAnnotation(a roadmap.Aged) string {
	src := "—"
	if a.Source != nil {
		src = a.Source.Feature
		if a.Source.Role != "" {
			src += "/" + a.Source.Role
		}
	}
	return fmt.Sprintf("%-28s %s", src, truncateSummary(a.Summary))
}

func truncateSummary(s string) string {
	r := []rune(strings.Join(strings.Fields(s), " "))
	if len(r) <= summaryWidth {
		return string(r)
	}
	return string(r[:summaryWidth-1]) + "…"
}
