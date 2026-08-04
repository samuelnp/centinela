package ui

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// RenderRoadmapSync renders the outcome of a post-mutation roadmap-state sync:
// what was committed, or why the state was left uncommitted. Presentation only
// — the decision of whether this is a warning lives in the report.
func RenderRoadmapSync(rep roadmap.SyncReport) string {
	if rep.Committed {
		return StyleGreen.Render(IconDone) + " " +
			fmt.Sprintf("Committed roadmap state: %s", rep.Message) +
			StyleMuted.Render("  ["+joinPaths(rep.Paths)+"]")
	}
	line := "Roadmap state left uncommitted"
	if rep.Reason != "" {
		line += ": " + rep.Reason
	}
	if rep.Warn {
		return StyleYellow.Render("⚠ " + line)
	}
	if rep.Regenerated {
		line += " (ROADMAP.md regenerated)"
	}
	return StyleMuted.Render(line)
}

// joinPaths renders a pathspec compactly for the trailing annotation.
func joinPaths(paths []string) string {
	out := ""
	for i, p := range paths {
		if i > 0 {
			out += " "
		}
		out += p
	}
	return out
}
