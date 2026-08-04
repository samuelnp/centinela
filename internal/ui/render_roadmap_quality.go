package ui

import "github.com/charmbracelet/lipgloss"

// RenderRoadmapQualityNeeded returns the setup panel directing the operator to
// delegate the roadmap quality-evaluator pass.
func RenderRoadmapQualityNeeded() string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⚠ Roadmap quality scoring missing — evaluator review required"),
		"",
		"Roadmap dependency analysis exists. Do not answer the user's message.",
		"Instead, delegate roadmap quality scoring to a roadmap quality evaluator.",
		"",
		StyleMuted.Render("1. Score each roadmap feature from 1-10 for:"),
		StyleMuted.Render("   acceptanceCriteria, userValue, definitionClarity, dependencies, effortEstimation"),
		StyleMuted.Render("2. Set an honest overall score per feature — it gates nothing"),
		StyleMuted.Render("3. Write .workflow/roadmap-quality.md summary and improvement loop"),
		StyleMuted.Render("4. Write .workflow/roadmap-quality.json with role roadmap-quality-evaluator"),
		StyleMuted.Render("   Include all roadmap features; every score must be 1-10"),
		StyleRed.Render("A low score blocks nothing — inflating one only hides work that needs doing."),
	)
	return renderSystemPanel("SETUP", "ROADMAP QUALITY REQUIRED", toneWarn, body)
}
