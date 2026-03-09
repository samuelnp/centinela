package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/samuelnp/centinela/internal/workflow"
)

// RenderBlocked returns a styled error box for the prewrite hook.
func RenderBlocked(fileType, step, feature, filePath string) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleRed.Render("✗  BLOCKED"),
		"",
		fmt.Sprintf("Can't write %q files during %q step.", fileType, step),
		"",
		StyleMuted.Render("Feature  ")+feature,
		StyleMuted.Render("File     ")+filePath,
	)
	return errorBoxStyle.Render(body)
}

// RenderTag returns a compact styled line for the postwrite hook.
func RenderTag(wf *workflow.Workflow) string {
	count := wfDoneCount(wf)
	return StyleMuted.Render(
		fmt.Sprintf("↳ %s · %s · %d/4", wf.Feature, wf.CurrentStep, count),
	)
}

// RenderContext returns a styled box summarising all active workflows,
// used by the UserPromptSubmit hook.
func RenderContext(wfs []*workflow.Workflow) string {
	var sections []string
	for _, wf := range wfs {
		count := wfDoneCount(wf)
		header := StyleBold.Render(wf.Feature) + "  " +
			StyleMuted.Render(fmt.Sprintf("%s %d/4", wf.CurrentStep, count))
		sections = append(sections, lipgloss.JoinVertical(lipgloss.Left, header, stepBar(wf)))
	}
	return boxStyle.Render(strings.Join(sections, "\n\n"))
}

func stepBar(wf *workflow.Workflow) string {
	parts := make([]string, 0, len(workflow.StepOrder))
	for _, step := range workflow.StepOrder {
		icon := stepIcon(wf, step)
		parts = append(parts, icon+" "+step)
	}
	return "  " + strings.Join(parts, "  ·  ")
}

func wfDoneCount(wf *workflow.Workflow) int {
	if wf.CurrentStep == "done" {
		return 4
	}
	for i, step := range workflow.StepOrder {
		if step == wf.CurrentStep {
			return i
		}
	}
	return 0
}

func stepIcon(wf *workflow.Workflow, step string) string {
	info := wf.Steps[step]
	switch {
	case step == wf.CurrentStep && info.Status != "done":
		return IconActive
	case info.Status == "done":
		return IconDone
	default:
		return IconPending
	}
}
