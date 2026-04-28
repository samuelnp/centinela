package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// RenderSetupNeeded returns context to inject when PROJECT.md is missing.
func RenderSetupNeeded() string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⚠ PROJECT.md not found — setup required"),
		"",
		"Do not answer the user's message. Instead, respond with:",
		"  \"This project needs to be configured before we can start.",
		"   Let me ask you a few questions to set it up.\"",
		"",
		"Then immediately:",
		StyleMuted.Render("1. Read PROJECT.md.template"),
		StyleMuted.Render("2. Ask these exact questions; do not combine or omit them:"),
		StyleMuted.Render("   1. Project name - what should we call it?"),
		StyleMuted.Render("   2. Elevator pitch - one sentence: what does it do and for whom?"),
		StyleMuted.Render("   3. Tech stack - language, framework, styling, persistence, test tools?"),
		StyleMuted.Render("   4. Architecture archetype - hexagonal, rails-native, n-tier, ecs, modular, or custom?"),
		StyleMuted.Render("   5. Locales - which languages does the UI support? (default: English only)"),
		StyleMuted.Render("   6. Folder layout - preferred structure, or should I propose one based on the archetype?"),
		StyleMuted.Render("3. Write PROJECT.md once you have all answers"),
		StyleMuted.Render("4. Tell the user: \"PROJECT.md is ready — next, let's define your roadmap.\""),
		StyleMuted.Render("   Then immediately start the roadmap conversation (phases, features, briefs)."),
		"",
		StyleRed.Render("Do not discuss anything else until PROJECT.md is written."),
	)
	return renderSystemPanel("SETUP", "PROJECT CONFIG REQUIRED", toneWarn, body)
}

// RenderProductionReadinessSetupNeeded returns context when the prompt file is missing.
func RenderProductionReadinessSetupNeeded() string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⚠ Production readiness prompt not configured"),
		"",
		"Do not answer the user's message. Instead:",
		StyleMuted.Render("1. Read PROJECT.md and"),
		StyleMuted.Render("   docs/architecture/production-readiness-prompt.md.template"),
		StyleMuted.Render("2. Ask the user about their key failure modes and external services"),
		StyleMuted.Render("3. Fill in [PLACEHOLDERS] with project-specific values"),
		StyleMuted.Render("4. Write docs/architecture/production-readiness-prompt.md"),
		"",
		StyleRed.Render("Do not continue until production-readiness-prompt.md is written."),
	)
	return renderSystemPanel("SETUP", "PRODUCTION READINESS REQUIRED", toneWarn, body)
}

// RenderProductionReadinessWarning returns a styled warning box for WARNING-status reports.
func RenderProductionReadinessWarning(feature string) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⚠ Production readiness: WARNING"),
		"",
		fmt.Sprintf("Non-critical issues found in %q.", feature),
		"Step advanced — but warnings should be addressed.",
		"",
		StyleMuted.Render("Suggested: centinela start "+feature+"-hardening"),
	)
	return renderSystemPanel("VALIDATE", "PRODUCTION READINESS WARNING", toneWarn, body)
}
