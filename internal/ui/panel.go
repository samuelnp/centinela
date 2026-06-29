package ui

import "github.com/charmbracelet/lipgloss"

type tone string

const (
	toneInfo    tone = "info"
	toneSuccess tone = "success"
	toneWarn    tone = "warn"
	toneError   tone = "error"
)

// renderSystemPanel renders a branded header line followed by the body, with no
// border box — the header (persona label + channel + title, tone-colored) plus
// the content below it is the whole rendering.
func renderSystemPanel(channel, title string, t tone, body string) string {
	return renderSystemLine(channel, title, t) + "\n\n" + body
}

func renderSystemLine(channel, title string, t tone) string {
	label := styleForTone(t).Bold(true).Render(personaLabel())
	meta := StyleMuted.Render(" " + channel + " ")
	text := StyleBold.Render(" " + title)
	return label + meta + text
}

func styleForTone(t tone) lipgloss.Style {
	if t == toneSuccess {
		return StyleGreen
	}
	if t == toneWarn {
		return StyleYellow
	}
	if t == toneError {
		return StyleRed
	}
	return StyleBlue
}
