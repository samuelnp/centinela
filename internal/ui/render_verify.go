package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/verify"
)

// RenderVerification renders a per-claim PASS/FAIL/SKIP/WARN/TIMEOUT report
// plus a summary line. Pure presentation: it makes no decisions, it only
// styles the result it is given.
func RenderVerification(r verify.VerificationResult) string {
	var lines []string
	lines = append(lines, StyleBold.Render("Claim Verification — "+r.Feature))
	for _, c := range r.Checks {
		lines = append(lines, renderCheckLine(c))
	}
	pass, fail, skip, warn := r.Tally()
	summary := fmt.Sprintf("%d passed, %d failed, %d warned, %d skipped", pass, fail, warn, skip)
	if fail > 0 {
		summary = StyleRed.Render(summary)
	} else if warn > 0 {
		summary = StyleYellow.Render(summary)
	} else {
		summary = StyleGreen.Render(summary)
	}
	lines = append(lines, "", summary)
	return strings.Join(lines, "\n")
}

func renderCheckLine(c verify.Check) string {
	label := string(c.Status)
	var styled string
	switch c.Status {
	case verify.StatusPass:
		styled = StyleGreen.Render(label)
	case verify.StatusFail, verify.StatusConfigError, verify.StatusTimeout:
		styled = StyleRed.Render(label)
	case verify.StatusWarn:
		styled = StyleYellow.Render(label)
	default:
		styled = StyleMuted.Render(label)
	}
	claim := c.Claim
	if c.Role != "" {
		claim += " (" + c.Role + ")"
	}
	return fmt.Sprintf("%s  %s  %s", styled, StyleBold.Render(claim), StyleMuted.Render(c.Detail))
}
