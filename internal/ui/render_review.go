package ui

import "github.com/charmbracelet/lipgloss"

// RenderReviewReady returns context to inject when a step's artifacts are complete
// but the step has not yet been advanced — reminding Claude to pause for user review.
func RenderReviewReady(feature, step, next string) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⏸  "+feature+" · "+step+" artifacts complete"),
		"",
		"STOP. Do not advance. Present what was written, then ask:",
		StyleMuted.Render("\"Step "+step+" is done — shall I advance to "+next+"?\""),
		"",
		StyleMuted.Render("Only run `centinela complete "+feature+"` after user confirms."),
	)
	return boxStyle.Render(body)
}
