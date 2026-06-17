package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// RenderPromoteEvaluatorContext returns the quality-evaluator context panel for
// `roadmap promote` without --scores. It lists the finding, the target phase,
// the threshold, the six scoring dimensions, and the literal re-invocation line.
// Pure formatting — promote writes nothing on this path.
func RenderPromoteEvaluatorContext(f *roadmap.BacklogFinding, phase string) string {
	src := "(none)"
	if f.Source != nil {
		src = f.Source.Feature
		if f.Source.Role != "" {
			src += "/" + f.Source.Role
		}
	}
	reinvoke := fmt.Sprintf(
		"centinela roadmap promote %s --phase %q --scores ac,uv,dc,dep,ee,overall", f.Name, phase)
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleBold.Render("Finding: ")+f.Name,
		StyleMuted.Render("Summary: ")+f.Summary,
		StyleMuted.Render("Source:  ")+src,
		StyleMuted.Render("Target phase: ")+phase,
		"",
		StyleBold.Render("Quality threshold: overall must be at least 9"),
		StyleMuted.Render("Score each dimension 1-10:"),
		StyleMuted.Render("  acceptanceCriteria, userValue, definitionClarity,"),
		StyleMuted.Render("  dependencies, effortEstimation, overall"),
		"",
		StyleBold.Render("Run an honest quality-evaluator pass, then re-invoke with --scores:"),
		StyleYellow.Render("  "+reinvoke),
	)
	return renderSystemPanel("ROADMAP", "QUALITY EVALUATOR CONTEXT", toneInfo, body)
}
