package planadvisor

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
)

func Directive(feature string, cfg *config.Config) string {
	if cfg == nil {
		cfg = &config.Config{}
	}
	mode := config.NormalizePlanAdvisorMode(cfg.Workflow.PlanAdvisorMode)
	if mode == config.PlanAdvisorOff {
		return ""
	}
	b := buildBundle(feature)
	questions := selectQuestions(b, config.NormalizePlanQuestionLimit(cfg.Workflow.PlanQuestionLimit), mode)
	lines := []string{
		fmt.Sprintf("CENTINELA PLAN ADVISOR: %q", feature),
		"Operate in two lenses: big-thinker first, then feature-specialist.",
		"Ask only the missing high-value questions from existing docs/features, docs/plans, specs, roadmap context, and prior edge-case lessons.",
		fmt.Sprintf("Ask at most %d questions this round. Do not jump to implementation.", config.NormalizePlanQuestionLimit(cfg.Workflow.PlanQuestionLimit)),
	}
	if ctx := contextLines(b); len(ctx) > 0 {
		lines = append(lines, "Relevant context:")
		lines = append(lines, ctx...)
	}
	if len(questions) == 0 {
		lines = append(lines, "Planning coverage looks solid. Synthesize the brief, plan, and spec without repeating generic discovery questions.")
		return strings.Join(lines, "\n")
	}
	for _, q := range questions {
		lines = append(lines, fmt.Sprintf("- [%s] %s", q.Lens, q.Text))
	}
	return strings.Join(lines, "\n")
}
