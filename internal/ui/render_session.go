package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// RenderSessionRehydration composes the SessionStart rehydration payload: a
// one-line banner, the full roadmap with per-feature readiness, the parallel
// frontier (all ready features), and a POINTERS block with read-on-demand paths.
// The ready set and hasIncomplete flag are computed by the caller, keeping this
// renderer pure and deterministic (no disk/workflow access in the UI layer).
func RenderSessionRehydration(r *roadmap.Roadmap, ready []string, hasIncomplete bool) string {
	readyBlock := renderReadyBlock(ready, hasIncomplete)
	pointers := []string{StyleMuted.Render("Pointers (read on demand):"),
		StyleMuted.Render("  · PROJECT.md")}
	for _, name := range ready {
		pointers = append(pointers, StyleMuted.Render("  · docs/features/"+name+".md"))
	}
	banner := StyleBold.Render("Session rehydration — project state recovered")
	body := lipgloss.JoinVertical(lipgloss.Left,
		banner,
		"",
		RenderRoadmap(r),
		"",
		readyBlock,
		"",
		lipgloss.JoinVertical(lipgloss.Left, pointers...),
	)
	return renderSystemPanel("SESSION", "REHYDRATION", toneInfo, body)
}

func renderReadyBlock(ready []string, hasIncomplete bool) string {
	if len(ready) > 0 {
		return StyleYellow.Render("Ready to start now:") + "\n" + RenderReadyList(ready)
	}
	if !hasIncomplete {
		return StyleGreen.Render("Roadmap complete — no next feature to plan.")
	}
	return StyleMuted.Render(
		"No features ready to start — everything in-progress or blocked by unmet dependencies.",
	)
}
