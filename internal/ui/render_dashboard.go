package ui

import (
	"strings"

	"github.com/samuelnp/centinela/internal/teamdashboard"
)

// RenderDashboard renders the multi-feature team-status board in house style:
// three sections — in-flight features, roadmap burn-down, and gate health —
// joined vertically. Each panel owns its empty state, so a board with no active
// work, no roadmap, and no gate failures still prints three honest panels.
// Output is deterministic and lipgloss auto-strips ANSI on non-TTY, so piped
// output is plain and parseable. Imports internal/teamdashboard read-only for
// the Dashboard type only (mirrors render_insights.go's insights import).
func RenderDashboard(d teamdashboard.Dashboard) string {
	parts := []string{
		dashHeader(d),
		featuresPanel(d.Features),
		roadmapPanel(d.Roadmap),
		gatesPanel(d.Gates),
	}
	return strings.Join(parts, "\n\n")
}

func dashHeader(d teamdashboard.Dashboard) string {
	return StyleBold.Render("Team Dashboard") + "  " +
		StyleMuted.Render(dashSummaryLabel(d))
}
