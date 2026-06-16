package planadvisor

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/insights"
	"github.com/samuelnp/centinela/internal/telemetry"
)

// recurringFailures returns the top-N most-recurring gate failures from the
// telemetry ledger, reusing the insights counter. Read-only and telemetry-gated:
// disabled telemetry or a missing/empty ledger yields nil, so the advisor stays
// byte-identical to its pre-feature output.
func recurringFailures(cfg *config.Config, topN int) []insights.Count {
	if cfg == nil || !cfg.Telemetry.IsEnabled() || topN <= 0 {
		return nil
	}
	events, err := telemetry.ReadDefault()
	if err != nil {
		return nil
	}
	return insights.Gates(events, topN)
}

func failureTopN(cfg *config.Config) int {
	if cfg == nil {
		return config.NormalizePlanAdvisorFailureTopN(0)
	}
	return config.NormalizePlanAdvisorFailureTopN(cfg.Workflow.PlanAdvisorFailureTopN)
}
