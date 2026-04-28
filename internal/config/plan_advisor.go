package config

import "strings"

const (
	PlanAdvisorOff         = "off"
	PlanAdvisorAlways      = "always"
	PlanAdvisorMissingInfo = "missing_info"
	DefaultPlanQuestionCap = 4
)

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
