package ui

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// RenderResolveSummary states the arithmetic of a resolved conflict. The
// per-side counts are the point: a human can check that nothing was dropped
// without re-reading the file (R5), which only works if they RECONCILE —
// unchanged + ours + theirs is exactly what was kept.
func RenderResolveSummary(m roadmap.Merged) string {
	return StyleGreen.Render(IconDone) + " " + fmt.Sprintf(
		"Resolved roadmap state — kept %d findings: %d unchanged, %d from our side, %d from theirs.",
		m.Kept, m.FromBase, m.FromOurs, m.FromTheirs)
}

// RenderResolveMarkdownOnly states the ROADMAP.md-only case: roadmap.json
// merged cleanly, so the generated file is simply regenerated from it.
func RenderResolveMarkdownOnly() string {
	return StyleGreen.Render(IconDone) + " " +
		"Regenerated ROADMAP.md from the merged roadmap.json and staged it."
}

// RenderResolveNothingToDo is the no-op state outside a conflict.
func RenderResolveNothingToDo() string {
	return StyleMuted.Render("Nothing to resolve — no roadmap-state path is conflicted.")
}
