package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/samuelnp/centinela/internal/workflow"
)

// RenderStatus returns a styled multi-line status view for a single workflow.
func RenderStatus(wf *workflow.Workflow) string {
	rows := []string{
		renderSystemLine("STATUS", "WORKFLOW", toneInfo),
		"",
		StyleBold.Render("Feature") + "  " + wf.Feature,
		StyleBold.Render("Started") + "  " + wf.StartedAt.Format("2006-01-02"),
		"",
	}
	for _, step := range wf.OrderedSteps() {
		info := wf.Steps[step]
		icon := stepIcon(wf, step)
		status := stepStatusLine(wf, step, info)
		rows = append(rows, fmt.Sprintf("  %s  %-10s  %s", icon, step, status))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// RenderSuccess returns a green check-mark line for confirmation messages.
func RenderSuccess(msg string) string {
	return renderSystemLine("CLI", msg, toneSuccess)
}

// RenderStep returns the step progress hint used after start/complete.
func RenderStep(label, step string) string {
	return renderSystemLine("CLI", label+": "+step, toneInfo)
}

func stepStatusLine(wf *workflow.Workflow, step string, info workflow.StepState) string {
	switch {
	case step == wf.CurrentStep && info.Status != "done":
		return StyleYellow.Render("in progress")
	case info.Status == "done":
		date := ""
		if info.CompletedAt != nil {
			date = "  " + StyleMuted.Render((*info.CompletedAt)[:10])
		}
		return StyleGreen.Render("done") + date
	default:
		return StyleMuted.Render("pending")
	}
}
