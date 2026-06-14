package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/doctor"
)

// RenderDiagnosis renders a single doctor diagnosis: a status glyph (✓/⚠/✗),
// the check name, the message, then indented detail lines and any report-only
// repair command. Color is applied via lipgloss styles, which auto-strip ANSI
// on non-TTY output so the report stays plain and parseable.
func RenderDiagnosis(d doctor.Diagnosis) string {
	lines := []string{glyphLine(d) + "  " + StyleMuted.Render(d.Message)}
	for _, det := range d.Details {
		lines = append(lines, StyleMuted.Render("  · "+det))
	}
	if d.Repair != nil && d.Repair.Command != "" {
		lines = append(lines, StyleMuted.Render("  → run: "+d.Repair.Command))
	}
	return strings.Join(lines, "\n")
}

// glyphLine renders the leading "<glyph> <name>" prefix in the status color.
func glyphLine(d doctor.Diagnosis) string {
	switch d.Status {
	case doctor.OK:
		return StyleGreen.Render("✓ " + d.Name)
	case doctor.Warn:
		return StyleYellow.Render("⚠ " + d.Name)
	default:
		return StyleRed.Render("✗ " + d.Name)
	}
}

// RenderDoctorSummary renders the trailing "N ok, M warn, K error" summary.
func RenderDoctorSummary(diags []doctor.Diagnosis) string {
	ok, warn, err := doctor.Counts(diags)
	return StyleBold.Render(fmt.Sprintf("%d ok, %d warn, %d error", ok, warn, err))
}
