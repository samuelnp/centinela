package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/audit"
)

// RenderAdoption renders the human adoption report in house style: a header, a
// per-gate accepted-violation count line, the total, and the ratchet-to-zero
// framing. For a zero-finding baseline it states there is nothing to ratchet.
// lipgloss auto-strips ANSI on non-TTY so piped output is plain.
func RenderAdoption(o audit.Outcome) string {
	total := o.Baseline.Total()
	header := StyleBold.Render(fmt.Sprintf(
		"Adopted baseline — %d accepted finding(s) across %d gate(s)",
		total, len(o.Baseline.Gates)))
	parts := []string{header}
	for _, e := range o.Baseline.Gates {
		parts = append(parts, StyleMuted.Render(
			fmt.Sprintf("  · %s: %d accepted finding(s)", e.Gate, len(e.Fingerprints))))
	}
	if total == 0 {
		parts = append(parts, StyleGreen.Render("0 accepted findings — nothing to ratchet."))
		return strings.Join(parts, "\n")
	}
	parts = append(parts, StyleGreen.Render(fmt.Sprintf(
		"Starting baseline: %d accepted finding(s) across %d gate(s) — ratchet to zero over time.",
		total, len(o.Baseline.Gates))))
	return strings.Join(parts, "\n")
}
