package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// RenderStatus returns a styled multi-line status view for a single workflow,
// resolving profile provenance without config (driver/global tiers unavailable).
func RenderStatus(wf *workflow.Workflow) string {
	return RenderStatusWithConfig(wf, nil)
}

// RenderStatusWithConfig renders the status view, threading cfg so the Profile
// row can show full provenance (driver model + which precedence tier applied).
func RenderStatusWithConfig(wf *workflow.Workflow, cfg *config.Config) string {
	rows := []string{
		renderSystemLine("STATUS", "WORKFLOW", toneInfo),
		"",
		StyleBold.Render("Feature") + "  " + wf.Feature,
		StyleBold.Render("Started") + "  " + wf.StartedAt.Format("2006-01-02"),
		StyleBold.Render("Profile") + profileLine(wf, cfg),
		StyleBold.Render("Archetype") + archetypeLine(wf),
	}
	if wf.WorktreePath != "" {
		rows = append(rows, StyleBold.Render("Worktree")+" "+wf.WorktreePath)
	}
	rows = append(rows, "")
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

// profileLine renders the effective profile plus its provenance annotation. The
// profile and note are computed in internal/workflow so the ui stays logic-free,
// matching the "Profile  <profile>  (<note>)" wording the status spec expects.
func profileLine(wf *workflow.Workflow, cfg *config.Config) string {
	profile, note := workflow.ProfileProvenance(wf, cfg)
	return "  " + profile + "  " + StyleMuted.Render("("+note+")")
}

// archetypeLine renders the archetype value plus the spike annotation, if any.
// The name and note are computed in internal/workflow so the ui stays logic-free.
func archetypeLine(wf *workflow.Workflow) string {
	name, note := workflow.DisplayArchetype(wf)
	if note != "" {
		return "  " + name + "  " + StyleMuted.Render("("+note+")")
	}
	return "  " + name
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
