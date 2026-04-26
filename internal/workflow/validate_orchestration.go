package workflow

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

func validateOrchestration(feature, step string, cfg *config.Config) error {
	if !strictOrchestrationEnabled(feature) {
		return nil
	}
	return orchestration.ValidateStep(feature, step, config.UIPaths(cfg))
}

func strictOrchestrationEnabled(feature string) bool {
	wf, err := Load(feature)
	if err != nil {
		return false
	}
	return wf.OrchestrationMode == StrictOrchestrationMode
}
