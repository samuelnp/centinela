package ui

import "github.com/charmbracelet/lipgloss"

func RenderRoadmapAnalysisNeeded() string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⚠ Roadmap analysis missing — senior PM review required"),
		"",
		"ROADMAP.md exists. Do not answer the user's message.",
		"Instead, delegate roadmap quality review to a senior product manager.",
		"",
		StyleMuted.Render("1. Analyze feature sequencing and UX flow continuity"),
		StyleMuted.Render("2. Validate cross-feature dependencies and prerequisites"),
		StyleMuted.Render("3. Write .workflow/roadmap-analysis.md summary"),
		StyleMuted.Render("4. Write .workflow/roadmap-analysis.json with role senior-product-manager"),
		StyleMuted.Render("   Include all roadmap features and dependsOn arrays"),
		"",
		StyleRed.Render("Do not start features until roadmap analysis artifacts are present."),
	)
	return renderSystemPanel("SETUP", "ROADMAP ANALYSIS REQUIRED", toneWarn, body)
}
