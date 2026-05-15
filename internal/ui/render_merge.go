package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// RenderMergeStewardNeeded returns the dispatch block shown when a merge
// stalls and the Merge Steward must resolve it. Read-only rendering.
func RenderMergeStewardNeeded(feature, reason string) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⚠ Merge Steward required — "+reason),
		"",
		"The merge of "+StyleBold.Render(feature)+" cannot complete on its own.",
		StyleMuted.Render("Delegate to the merge-steward subagent using"),
		StyleMuted.Render("docs/architecture/merge-steward-prompt.md, then run:"),
		StyleMuted.Render("  centinela merge --continue "+feature),
	)
	return renderSystemPanel("MERGE", "STEWARD REQUIRED", toneWarn, body)
}

// RenderMergeEscalated returns the block shown when steward evidence
// escalates: the merge stays blocked and the worktree is kept.
func RenderMergeEscalated(feature string) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleRed.Render("✗ Merge Steward escalated — not finalized"),
		"",
		"The Steward could not resolve "+StyleBold.Render(feature)+" with high confidence.",
		StyleMuted.Render("Worktree and pending marker kept for manual review."),
		StyleMuted.Render("Escalation note and proposed diff follow:"),
	)
	return renderSystemPanel("MERGE", "STEWARD ESCALATED", toneError, body)
}
