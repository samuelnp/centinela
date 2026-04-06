package main

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

func shouldRenderReviewPrompt(wf *workflow.Workflow, cfg *config.Config) bool {
	if wf == nil || wf.CurrentStep == "done" {
		return false
	}
	mode := config.NormalizeStepConfirmationMode(cfg.Workflow.StepConfirmationMode)
	if mode == config.ConfirmAuto {
		return false
	}
	if mode == config.ConfirmAfterPlan {
		return wf.CurrentStep == "plan"
	}
	return true
}
