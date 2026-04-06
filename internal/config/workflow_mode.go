package config

import "strings"

const (
	ConfirmEveryStep = "every_step"
	ConfirmAfterPlan = "after_plan"
	ConfirmAuto      = "auto"
)

func NormalizeStepConfirmationMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case ConfirmAfterPlan:
		return ConfirmAfterPlan
	case ConfirmAuto:
		return ConfirmAuto
	default:
		return ConfirmEveryStep
	}
}
