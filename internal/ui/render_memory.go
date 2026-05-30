package ui

import "strings"

// RenderMemoryBlock renders recalled ledger facts as a compact terminal block.
// Input is pre-formatted plain-text summaries; rendering carries no logic.
func RenderMemoryBlock(facts []string) string {
	if len(facts) == 0 {
		return ""
	}
	lines := []string{StyleBlue.Bold(true).Render("🛡️👁️ MEMORY")}
	for _, f := range facts {
		lines = append(lines, StyleMuted.Render("  · ")+f)
	}
	return strings.Join(lines, "\n")
}
