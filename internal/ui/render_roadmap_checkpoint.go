package ui

import "github.com/charmbracelet/lipgloss"

// RenderRoadmapCheckpoint returns the system panel asking the user whether to
// keep iterating on the roadmap or start the first incomplete Phase 0 feature.
// featureName is the resolved first incomplete bootstrap feature. This is pure
// formatting — all emit/suppress decisions live in internal/roadmapcheckpoint.
func RenderRoadmapCheckpoint(featureName string) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleBold.Render("Roadmap definition iteration complete."),
		"",
		"Do not answer the user's message yet. Ask the user to choose:",
		"",
		StyleBold.Render("Option A — keep iterating on the roadmap"),
		StyleMuted.Render("   Refine ROADMAP.md, analysis, or quality artifacts."),
		StyleMuted.Render("   If the user picks this, run:"),
		StyleYellow.Render("   centinela roadmap iterate"),
		StyleMuted.Render("   to persist a marker that suppresses this prompt until"),
		StyleMuted.Render("   a roadmap-defining artifact changes again."),
		"",
		StyleBold.Render("Option B — start the first Phase 0 feature"),
		StyleMuted.Render("   Begin implementing:")+" "+StyleBold.Render(featureName),
		StyleMuted.Render("   If the user picks this, run:"),
		StyleYellow.Render("   centinela start "+featureName),
		"",
		StyleMuted.Render("Editing any roadmap artifact later re-fires this prompt."),
		"",
		StyleRed.Render("Wait for the user's choice before proceeding."),
	)
	return renderSystemPanel("ROADMAP", "CHECKPOINT", toneInfo, body)
}
