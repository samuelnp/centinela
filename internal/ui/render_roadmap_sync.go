package ui

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// RenderRoadmapSync renders the outcome of a post-mutation roadmap-state sync.
//
// It never claims where the record is without evidence. "Left uncommitted"
// reads as "it is on disk, commit it yourself", so that phrasing is only used
// when SyncReport.InWorkingTree — a verified read-back — supports it. When the
// read-back fails the line says the opposite, loudly: a lost record that
// reports as merely uncommitted is invisible, because this feature's own
// auto-commit leaves a clean tree and `git status` will not show it either.
func RenderRoadmapSync(rep roadmap.SyncReport) string {
	if rep.Committed {
		return StyleGreen.Render(IconDone) + " " +
			fmt.Sprintf("Committed roadmap state: %s", rep.Message) +
			StyleMuted.Render("  ["+joinPaths(rep.Paths)+"]")
	}
	if !rep.InWorkingTree {
		return StyleRed.Render("✖ " + unverifiedLine(rep.Reason))
	}
	line := "Roadmap state was not committed; it is in your working tree"
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

// unverifiedLine is the honest report when the read-back failed: state what
// this command did NOT do, refuse to say where the record is, and name the
// file to re-check.
func unverifiedLine(reason string) string {
	line := "Roadmap state was NOT committed, and roadmap.json no longer matches " +
		"what this command wrote — another process may have overwritten it"
	if reason != "" {
		line += " (" + reason + ")"
	}
	return line + ". Re-check .workflow/roadmap.json before retrying."
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
