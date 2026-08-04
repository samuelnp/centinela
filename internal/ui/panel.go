package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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

// trimTrailingWS strips trailing spaces/tabs from every line of a hook-only
// panel render. lipgloss.JoinVertical(lipgloss.Left, ...) right-pads every
// line in a block to the width of that block's longest line — a leftover of
// the fixed-width bordered-box era (removed in remove-panel-borders,
// 3ab3cef) with no remaining visual purpose, since these panels render as
// flat text with no rectangle left to square.
func trimTrailingWS(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return strings.Join(lines, "\n")
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
