package ui

import (
	"strings"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// renderBacklogSection returns the Backlog findings block for the roadmap panel,
// or "" when the Backlog phase is missing or empty. Each finding is shown as
// "○ <slug>  <summary>" with the summary muted — no readiness state applies.
func renderBacklogSection(r *roadmap.Roadmap) string {
	findings := roadmap.BacklogFeatures(r)
	if len(findings) == 0 {
		return ""
	}
	lines := []string{StyleBold.Render(roadmap.BacklogPhaseName)}
	for _, f := range findings {
		lines = append(lines, "  "+IconPending+" "+f.Name+
			StyleMuted.Render("  "+f.Summary))
	}
	return strings.Join(lines, "\n")
}
