package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/samuelnp/centinela/internal/migration"
)

func RenderDocsMigrationPlan(plan migration.Plan, applying bool) string {
	mode := "preview"
	if applying {
		mode = "apply"
	}
	var lines []string
	for _, it := range plan.Items {
		line := fmt.Sprintf("- %s: %s (%s -> %s)", it.Action, it.Path, it.FromVersion, it.ToVersion)
		if it.PreservedKeepBlocks > 0 || it.PreservedCustomSection > 0 {
			line += fmt.Sprintf(" [keep:%d custom:%d]", it.PreservedKeepBlocks, it.PreservedCustomSection)
		}
		lines = append(lines, line)
	}
	body := lipgloss.JoinVertical(lipgloss.Left,
		fmt.Sprintf("Managed docs requiring migration: %d", len(plan.Items)),
		strings.Join(lines, "\n"),
	)
	return renderSystemPanel("MIGRATE", "DOCS "+strings.ToUpper(mode), toneInfo, body)
}

func RenderDocsMigrationNeeded(plan migration.Plan) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⚠ Managed docs are outdated"),
		fmt.Sprintf("Detected %d file(s) requiring migration.", len(plan.Items)),
		StyleMuted.Render("Ask the user for approval before applying changes."),
		StyleMuted.Render("Preview: centinela migrate docs"),
		StyleMuted.Render("Apply:   centinela migrate docs --apply"),
	)
	return renderSystemPanel("SETUP", "DOC MIGRATION REQUIRED", toneWarn, body)
}
