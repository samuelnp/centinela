package workflow

import "github.com/samuelnp/centinela/internal/orchestration"

func validateOrchestration(feature, step string) error {
	if !strictOrchestrationEnabled(feature) {
		return nil
	}
	return orchestration.ValidateStep(feature, step)
}

func strictOrchestrationEnabled(feature string) bool {
	wf, err := Load(feature)
	if err != nil {
		return false
	}
	return wf.OrchestrationMode == StrictOrchestrationMode
}
