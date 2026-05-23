package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// RenderSessionRehydration composes the SessionStart rehydration payload: a
// one-line banner, the full roadmap with per-feature status, the next feature
// to plan (or a roadmap-complete line when !hasNext), and a POINTERS block
// listing read-on-demand file PATHS (never inlined content). Pure presentation.
func RenderSessionRehydration(r *roadmap.Roadmap, next string, hasNext bool) string {
	banner := StyleBold.Render("Session rehydration — project state recovered")
	var nextLine string
	pointers := []string{StyleMuted.Render("Pointers (read on demand):"),
		StyleMuted.Render("  · PROJECT.md")}
	if hasNext {
		nextLine = StyleYellow.Render("Next feature to plan: " + next)
		pointers = append(pointers, StyleMuted.Render("  · docs/features/"+next+".md"))
	} else {
		nextLine = StyleGreen.Render("Roadmap complete — no next feature to plan.")
	}
	body := lipgloss.JoinVertical(lipgloss.Left,
		banner,
		"",
		RenderRoadmap(r),
		"",
		nextLine,
		"",
		lipgloss.JoinVertical(lipgloss.Left, pointers...),
	)
	return renderSystemPanel("SESSION", "REHYDRATION", toneInfo, body)
}
