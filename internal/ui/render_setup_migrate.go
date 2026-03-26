package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/samuelnp/centinela/internal/setup"
)

func RenderSetupMigrationPlan(plan setup.SyncPlan, applying bool) string {
	mode := "preview"
	if applying {
		mode = "apply"
	}
	lines := []string{}
	for _, it := range plan.Items {
		line := fmt.Sprintf("- %s: %s", it.Action, it.Path)
		if it.Reason != "" {
			line += " (" + it.Reason + ")"
		}
		lines = append(lines, line)
	}
	body := lipgloss.JoinVertical(lipgloss.Left,
		fmt.Sprintf("Managed setup assets requiring migration: %d", len(plan.Items)),
		strings.Join(lines, "\n"),
	)
	return renderSystemPanel("MIGRATE", "SETUP "+strings.ToUpper(mode), toneInfo, body)
}

func RenderMigrationNeeded(docsCount, setupCount int) string {
	parts := []string{}
	if docsCount > 0 {
		parts = append(parts, fmt.Sprintf("docs:%d", docsCount))
	}
	if setupCount > 0 {
		parts = append(parts, fmt.Sprintf("setup:%d", setupCount))
	}
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⚠ Managed migration available"),
		"Detected changes for "+strings.Join(parts, " "),
		StyleMuted.Render("Ask the user for approval before applying changes."),
		StyleMuted.Render("Preview: centinela migrate"),
		StyleMuted.Render("Apply:   centinela migrate --apply"),
	)
	return renderSystemPanel("SETUP", "MIGRATION REQUIRED", toneWarn, body)
}
