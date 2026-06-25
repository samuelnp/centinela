package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/samuelnp/centinela/internal/gitutil"
)

// RenderDeliveryChoice returns the panel shown at completion that lists only
// the delivery options valid for this repo and the exact command for each.
// Read-only presentation; it performs no delivery and has no side effects.
func RenderDeliveryChoice(feature string, opts []gitutil.Option) string {
	if len(opts) == 0 {
		body := lipgloss.JoinVertical(lipgloss.Left,
			StyleYellow.Render("⚠ No delivery target detected for "+StyleBold.Render(feature)),
			"",
			StyleMuted.Render("Configure an `origin` remote or run in worktree mode,"),
			StyleMuted.Render("then deliver this feature."),
		)
		return renderSystemPanel("DELIVER", "CHOOSE DELIVERY", toneWarn, body)
	}
	lines := []string{
		"Completed " + StyleBold.Render(feature) + " — ask the user how to deliver it.",
		StyleMuted.Render("Run only the option they pick:"),
		"",
	}
	for _, o := range opts {
		lines = append(lines, "  "+StyleGreen.Render("centinela deliver "+feature+" --via "+string(o)))
	}
	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return renderSystemPanel("DELIVER", "CHOOSE DELIVERY", toneInfo, body)
}
