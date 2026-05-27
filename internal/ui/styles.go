package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorGreen  = lipgloss.Color("2")
	colorRed    = lipgloss.Color("1")
	colorYellow = lipgloss.Color("3")
	colorGray   = lipgloss.Color("240")
	colorBlue   = lipgloss.Color("69")

	StyleBold   = lipgloss.NewStyle().Bold(true)
	StyleMuted  = lipgloss.NewStyle().Foreground(colorGray)
	StyleGreen  = lipgloss.NewStyle().Foreground(colorGreen)
	StyleRed    = lipgloss.NewStyle().Bold(true).Foreground(colorRed)
	StyleYellow = lipgloss.NewStyle().Foreground(colorYellow)
	StyleBlue   = lipgloss.NewStyle().Foreground(colorBlue)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBlue).
			Padding(0, 1)

	errorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorRed).
			Padding(0, 1)

	IconActive  = StyleYellow.Render("▶")
	IconDone    = StyleGreen.Render("✓")
	IconPending = StyleMuted.Render("○")
	IconReady   = "🔓"
	IconBlocked = "🔒"
)
