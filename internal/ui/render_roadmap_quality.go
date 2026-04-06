package ui

import "github.com/charmbracelet/lipgloss"

func RenderRoadmapQualityNeeded() string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⚠ Roadmap quality scoring missing — evaluator review required"),
		"",
		"Roadmap dependency analysis exists. Do not answer the user's message.",
		"Instead, delegate roadmap quality scoring to a roadmap quality evaluator.",
		"",
		StyleMuted.Render("1. Score each roadmap feature from 1-10 for:"),
		StyleMuted.Render("   acceptanceCriteria, userValue, definitionClarity, dependencies, effortEstimation"),
		StyleMuted.Render("2. Set overall score per feature (overall is the gate)"),
		StyleMuted.Render("3. Write .workflow/roadmap-quality.md summary and improvement loop"),
		StyleMuted.Render("4. Write .workflow/roadmap-quality.json with role roadmap-quality-evaluator"),
		StyleMuted.Render("   Set threshold to 9 and include all roadmap features"),
		StyleRed.Render("Iterate roadmap and feature briefs until every feature overall score is >= 9."),
	)
	return renderSystemPanel("SETUP", "ROADMAP QUALITY REQUIRED", toneWarn, body)
}
