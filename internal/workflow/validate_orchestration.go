package workflow

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

func validateOrchestration(feature, step string, cfg *config.Config) error {
	if !strictOrchestrationEnabled(feature) {
		return nil
	}
	return orchestration.ValidateRoles(feature, step, requiredRoles(feature, step), config.UIPaths(cfg))
}

// requiredRoles resolves the evidence roles for a step. The validate step is
// the one place the pinned contract overrides policy: a workflow started
// before adversarial-v1 keeps demanding validation-specialist evidence, which
// is what it actually wrote, and is never retro-gated on a verifier report.
func requiredRoles(feature, step string) []orchestration.Role {
	if step == "validate" && !featureUsesAdversarialVerifier(feature) {
		return []orchestration.Role{orchestration.RoleValidationSpec}
	}
	return orchestration.RequiredRolesForFeature(feature, step)
}

func strictOrchestrationEnabled(feature string) bool {
	wf, err := Load(feature)
	if err != nil {
		return false
	}
	return wf.OrchestrationMode == StrictOrchestrationMode
}
