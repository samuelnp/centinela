package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/cost"
)

// RenderCost renders the cost-governance report in house style: one section per
// scope (feature, step, model), each row showing used/budget/remaining with an
// over-budget marker. An empty report renders a single muted line. lipgloss
// strips ANSI on non-TTY, so piped output is plain and parseable.
func RenderCost(r cost.Report) string {
	if r.Empty() {
		return StyleMuted.Render("no cost samples yet — run governed workflows with [cost] enabled")
	}
	parts := []string{
		StyleBold.Render("Cost — token spend vs budget"),
		costSection("By feature", r.Features),
		costSection("By step", r.Steps),
		costSection("By model", r.Models),
	}
	return strings.Join(parts, "\n\n")
}

// RenderCostWarning is the single non-failing soft-gate line for an over-budget
// active scope, surfaced by validate and the status line.
func RenderCostWarning(s cost.Status) string {
	return StyleYellow.Render(fmt.Sprintf("⚠ cost: %s %q over budget — %d / %d tokens",
		s.Scope, s.Name, s.Used, s.Budget))
}

func costSection(title string, rows []cost.Status) string {
	if len(rows) == 0 {
		return StyleBold.Render(title) + "\n" + StyleMuted.Render("  (none)")
	}
	lines := []string{StyleBold.Render(title)}
	for _, s := range rows {
		lines = append(lines, "  "+costRow(s))
	}
	return strings.Join(lines, "\n")
}

func costRow(s cost.Status) string {
	if s.Budget <= 0 {
		return fmt.Sprintf("%s  %d tokens", s.Name, s.Used)
	}
	body := fmt.Sprintf("%s  %d / %d tokens  (%d left)", s.Name, s.Used, s.Budget, s.Remaining())
	if s.Over {
		return StyleYellow.Render("⚠ " + body + "  OVER")
	}
	return StyleGreen.Render(body)
}
