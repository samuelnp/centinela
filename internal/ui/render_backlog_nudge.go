package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// RenderBacklogNudge renders the completion prompt: the roadmap is finished but
// the Backlog is not, with the two commands that close it. It is guidance only
// — `complete` prints it and moves on.
func RenderBacklogNudge(n roadmap.Nudge) string {
	return strings.Join([]string{
		StyleYellow.Render(fmt.Sprintf(
			"Roadmap complete — %d deferred findings remain in the Backlog (%d older than %dd).",
			n.Total, n.Stale, n.ThresholdDays)),
		StyleMuted.Render("  Review:   centinela roadmap backlog --stale"),
		StyleMuted.Render(`  Schedule: centinela roadmap promote <slug> --phase "<phase>"`),
	}, "\n")
}
