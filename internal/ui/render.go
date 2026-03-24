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
		StyleRed.Render("✖ Write blocked by workflow policy"),
		"",
		fmt.Sprintf("Can't write %q files during %q step.", fileType, step),
		"",
		StyleMuted.Render("Feature  ")+feature,
		StyleMuted.Render("File     ")+filePath,
		"",
		StyleYellow.Render("Next action: "+blockedAction(step, feature)),
	)
	return renderSystemPanel("HOOK", "BLOCKED WRITE", toneError, body)
}

// RenderTag returns a compact styled line for the postwrite hook.
func RenderTag(wf *workflow.Workflow) string {
	count := wfDoneCount(wf)
	total := len(wf.OrderedSteps())
	line := fmt.Sprintf("feature=%s  step=%s  progress=%d/%d", wf.Feature, wf.CurrentStep, count, total)
	return renderSystemLine("HOOK", "TAG  "+line, toneInfo)
}

// RenderContext returns a styled box summarising all active workflows,
// used by the UserPromptSubmit hook.
func RenderContext(wfs []*workflow.Workflow) string {
	var sections []string
	for _, wf := range wfs {
		count := wfDoneCount(wf)
		total := len(wf.OrderedSteps())
		header := StyleBold.Render(wf.Feature) + "  " +
			StyleMuted.Render(fmt.Sprintf("%s %d/%d", wf.CurrentStep, count, total))
		sections = append(sections, lipgloss.JoinVertical(lipgloss.Left, header, stepBar(wf)))
	}
	body := strings.Join(sections, "\n\n")
	return renderSystemPanel("HOOK", "ACTIVE WORKFLOWS", toneInfo, body)
}

func stepBar(wf *workflow.Workflow) string {
	steps := wf.OrderedSteps()
	parts := make([]string, 0, len(steps))
	for _, step := range steps {
		icon := stepIcon(wf, step)
		parts = append(parts, icon+" "+step)
	}
	return "  " + strings.Join(parts, "  ·  ")
}

func wfDoneCount(wf *workflow.Workflow) int {
	steps := wf.OrderedSteps()
	if wf.CurrentStep == "done" {
		return len(steps)
	}
	for i, step := range steps {
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

func blockedAction(step, feature string) string {
	if step == "" || feature == "" || feature == "—" {
		return "run `centinela start <feature>` to create a workflow"
	}
	return "finish current step and run `centinela complete " + feature + "`"
}
