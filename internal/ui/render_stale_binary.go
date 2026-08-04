package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// RenderBlockedStaleBinary is the refusal for a write governed only by state
// files written by a NEWER Centinela than this binary.
//
// It exists rather than reusing RenderBlocked because that panel's "Next
// action" is always about the workflow's step — "finish current step and run
// centinela complete <feature>" — and here that is advice the tool itself will
// refuse: Save fails closed on a future version, so completing is exactly what
// cannot happen until the binary is upgraded. A refusal whose headline names an
// impossible remedy is worse than a terse one.
func RenderBlockedStaleBinary(fileType, feature, filePath string) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleRed.Render("✖ Write blocked: this Centinela is too old to read the workflow"),
		"",
		fmt.Sprintf("%q state was written by a newer Centinela, so its step rules cannot be read", feature),
		"and guessing at them would either block every write or wave them all through.",
		"",
		StyleMuted.Render("Feature  ")+feature,
		StyleMuted.Render("File     ")+filePath,
		StyleMuted.Render("Type     ")+fileType,
		"",
		StyleYellow.Render("Next action: run `centinela update`, then retry the write"),
	)
	return renderSystemPanel("HOOK", "BLOCKED WRITE", toneError, body)
}
