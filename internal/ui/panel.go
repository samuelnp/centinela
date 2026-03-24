package ui

import "github.com/charmbracelet/lipgloss"

type tone string

const (
	toneInfo    tone = "info"
	toneSuccess tone = "success"
	toneWarn    tone = "warn"
	toneError   tone = "error"
)

func renderSystemPanel(channel, title string, t tone, body string) string {
	head := renderSystemLine(channel, title, t)
	content := lipgloss.JoinVertical(lipgloss.Left, head, "", body)
	return panelStyle(t).Render(content)
}

func renderSystemLine(channel, title string, t tone) string {
	label := styleForTone(t).Bold(true).Render(" CENTINELA ")
	meta := StyleMuted.Render(" " + channel + " ")
	text := StyleBold.Render(" " + title)
	return label + meta + text
}

func panelStyle(t tone) lipgloss.Style {
	s := boxStyle
	s = s.BorderForeground(colorBlue)
	if t == toneSuccess {
		s = s.BorderForeground(colorGreen)
	}
	if t == toneWarn {
		s = s.BorderForeground(colorYellow)
	}
	if t == toneError {
		s = s.BorderForeground(colorRed)
	}
	return s
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
