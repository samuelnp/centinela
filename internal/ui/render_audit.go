package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/audit"
)

// RenderAuditDiff renders a ratchet diff in house style: a header with the
// new/baselined/resolved counts, then the new violations (the blocking set) and
// the prunable resolved set. lipgloss auto-strips ANSI on non-TTY so piped
// output is plain and parseable.
func RenderAuditDiff(d audit.Diff) string {
	header := StyleBold.Render(fmt.Sprintf(
		"Audit — %d new, %d baselined, %d resolved",
		len(d.New), len(d.Baselined), len(d.Resolved)))
	parts := []string{header}
	if s := auditSection("New (blocking)", StyleRed, d.New); s != "" {
		parts = append(parts, s)
	}
	if s := auditSection("Resolved (prunable on next baseline)", StyleGreen, d.Resolved); s != "" {
		parts = append(parts, s)
	}
	if !d.HasNew() {
		parts = append(parts, StyleGreen.Render("✓ no new violations since baseline"))
	}
	return strings.Join(parts, "\n\n")
}

func auditSection(title string, style interface{ Render(...string) string }, fps []audit.Fingerprint) string {
	if len(fps) == 0 {
		return ""
	}
	lines := []string{style.Render(title)}
	for _, fp := range fps {
		lines = append(lines, StyleMuted.Render("  · "+fp.Gate+": "+fp.Raw))
	}
	return strings.Join(lines, "\n")
}
