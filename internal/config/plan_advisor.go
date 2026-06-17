package config

import "strings"

const (
	PlanAdvisorOff         = "off"
	PlanAdvisorAlways      = "always"
	PlanAdvisorMissingInfo = "missing_info"
	DefaultPlanQuestionCap = 4

	DefaultPlanAdvisorFailureTopN = 3
	MaxPlanAdvisorFailureTopN     = 5
)

// NormalizePlanAdvisorFailureTopN clamps the recurring-failure list size to
// [1, MaxPlanAdvisorFailureTopN], defaulting an unset (<=0) value to 3.
func NormalizePlanAdvisorFailureTopN(n int) int {
	if n <= 0 {
		return DefaultPlanAdvisorFailureTopN
	}
	if n > MaxPlanAdvisorFailureTopN {
		return MaxPlanAdvisorFailureTopN
	}
	return n
}

func NormalizePlanAdvisorMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case PlanAdvisorOff:
		return PlanAdvisorOff
	case PlanAdvisorAlways:
		return PlanAdvisorAlways
	default:
		return PlanAdvisorMissingInfo
	}
}

func NormalizePlanQuestionLimit(limit int) int {
	if limit <= 0 || limit > DefaultPlanQuestionCap {
		return DefaultPlanQuestionCap
	}
	return limit
}
