package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// RenderFeatureBriefNeeded fires when a workflow is at "plan" but no feature brief exists yet.
func RenderFeatureBriefNeeded(feature string) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⚠ Feature brief missing: "+feature),
		"",
		"Before writing the plan, document this feature deeply.",
		"Ask the user the questions below, then write:",
		StyleBold.Render("  docs/features/"+feature+".md"),
		"",
		StyleMuted.Render("## Problem — what pain does this solve? Who is the user?"),
		StyleMuted.Render("## User Stories — as a [user], I want [x] so that [y]"),
		StyleMuted.Render("## Acceptance Criteria — concrete, testable (-> Gherkin scenarios)"),
		StyleMuted.Render("## Edge Cases — invalid input, concurrency, empty state, limits"),
		StyleMuted.Render("## Data Model — entities, key fields, relationships"),
		StyleMuted.Render("## Integration Points — APIs, events, external services"),
		StyleMuted.Render("## Risks — performance, security, unclear requirements"),
		StyleMuted.Render("## Decomposition — if large, list sub-feature slugs to split into"),
		"",
		StyleRed.Render(fmt.Sprintf("Write the brief first. The plan must reference it: docs/features/%s.md", feature)),
	)
	return renderSystemPanel("HOOK", "FEATURE BRIEF REQUIRED", toneWarn, body)
}
