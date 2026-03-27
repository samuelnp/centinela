package ui

import "github.com/charmbracelet/lipgloss"

// RenderReviewReady returns context to inject when a step's artifacts are complete
// but the step has not yet been advanced — reminding Claude to pause for user review.
func RenderReviewReady(feature, step, next string) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⏸ "+feature+" · "+step+" artifacts complete"),
		"",
		"STOP. Do not advance. Present what was written, then ask:",
		StyleMuted.Render("\"Step "+step+" is done — shall I advance to "+next+"?\""),
		"",
		StyleMuted.Render("Only run `centinela complete "+feature+"` after user confirms."),
	)
	return renderSystemPanel("HOOK", "REVIEW REQUIRED", toneWarn, body)
}

// RenderEdgeCaseReportNeeded reminds the agent to write edge-case analysis.
func RenderEdgeCaseReportNeeded(feature string) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⚠ Edge-case report missing: "+feature),
		"",
		"Tests phase requires hard-path analysis before completion.",
		StyleMuted.Render("Run edge-case subagent using docs/architecture/edge-case-tester-prompt.md"),
		StyleMuted.Render("Then write: .workflow/"+feature+"-edge-cases.md"),
	)
	return renderSystemPanel("HOOK", "ACTION REQUIRED", toneWarn, body)
}

// RenderDocumentationNeeded reminds the agent to run docs-specialist workflow.
func RenderDocumentationNeeded(feature string) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⚠ Documentation output missing: "+feature),
		"",
		"Docs step requires generated human-facing documentation.",
		StyleMuted.Render("Run docs specialist using docs/architecture/documentation-generator-prompt.md"),
		StyleMuted.Render("Then write: docs/project-docs/index.html"),
	)
	return renderSystemPanel("HOOK", "ACTION REQUIRED", toneWarn, body)
}
