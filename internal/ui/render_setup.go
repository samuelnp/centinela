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

// RenderBrownfieldSetupNeeded returns context to inject when PROJECT.md is
// missing but the repo already contains source. Instead of cold-interrogating
// the user, the agent drafts PROJECT.md from the codebase, then confirms.
func RenderBrownfieldSetupNeeded() string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⚠ Existing code detected — PROJECT.md missing"),
		"",
		"Do NOT interrogate the user with setup questions. This repo already",
		"has source code. Draft PROJECT.md from the codebase, then confirm.",
		"",
		StyleMuted.Render("1. Run `centinela analyze` (scans the repo into .workflow/analysis.json)"),
		StyleMuted.Render("2. Run `centinela synthesize` (drafts PROJECT.md from the inventory;"),
		StyleMuted.Render("   infers the archetype)"),
		StyleMuted.Render("3. ENRICH the draft — read the key source (design docs, manifests like"),
		StyleMuted.Render("   package.json/go.mod, i18n locale files) to correct inferred guesses"),
		StyleMuted.Render("   and fill gaps"),
		StyleMuted.Render("4. Set `**Project Stage:** existing` in PROJECT.md (so the workflow"),
		StyleMuted.Render("   skips greenfield bootstrap)"),
		StyleMuted.Render("5. Present the drafted PROJECT.md to the user, confirm any uncertain"),
		StyleMuted.Render("   fields, THEN finalize"),
		"",
		StyleRed.Render("Draft from the code first — confirm with the user — then write PROJECT.md."),
	)
	return renderSystemPanel("SETUP", "BROWNFIELD PROJECT DETECTED", toneWarn, body)
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
