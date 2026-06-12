package main

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

func shouldRenderReviewPrompt(wf *workflow.Workflow, cfg *config.Config) bool {
	if wf == nil || wf.CurrentStep == "done" {
		return false
	}
	mode := effectiveConfirmationMode(wf, cfg)
	if mode == config.ConfirmAuto {
		return false
	}
	if mode == config.ConfirmAfterPlan {
		return wf.CurrentStep == "plan"
	}
	return true
}

// effectiveConfirmationMode resolves step_confirmation_mode by precedence: an
// explicit raw knob in centinela.toml wins, else the effective profile's
// default, else the hardcoded every_step. The raw value (pre-normalization)
// distinguishes an explicit every_step from a defaulted one.
func effectiveConfirmationMode(wf *workflow.Workflow, cfg *config.Config) string {
	if cfg != nil && cfg.Workflow.RawStepConfirmationMode != "" {
		return config.NormalizeStepConfirmationMode(cfg.Workflow.RawStepConfirmationMode)
	}
	return config.ProfileDefaults(workflow.EffectiveProfile(wf, cfg)).ConfirmationMode
}
