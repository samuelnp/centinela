package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// epConfirmMode mirrors the cmd-layer resolver: explicit raw knob wins, else the
// effective profile default. Asserting it here exercises the same precedence the
// review-prompt decision uses, without importing package main.
func epConfirmMode(wf *workflow.Workflow, cfg *config.Config) string {
	if cfg != nil && cfg.Workflow.RawStepConfirmationMode != "" {
		return config.NormalizeStepConfirmationMode(cfg.Workflow.RawStepConfirmationMode)
	}
	return config.ProfileDefaults(workflow.EffectiveProfile(wf, cfg)).ConfirmationMode
}

func epRenders(wf *workflow.Workflow, cfg *config.Config) bool {
	switch epConfirmMode(wf, cfg) {
	case config.ConfirmAuto:
		return false
	case config.ConfirmAfterPlan:
		return wf.CurrentStep == "plan"
	default:
		return true
	}
}

// Acceptance: specs/enforcement-profiles.feature

// Scenario: Outcome profile suppresses the stop-and-ask review prompt
func TestEP_OutcomeSuppressesReviewPrompt(t *testing.T) {
	cfg := &config.Config{}
	cfg.Workflow.EnforcementProfile = config.ProfileOutcome
	for _, step := range []string{"plan", "code", "tests", "validate", "docs"} {
		if epRenders(&workflow.Workflow{CurrentStep: step}, cfg) {
			t.Fatalf("outcome must suppress the review prompt at step %q", step)
		}
	}
}

// Scenario: An explicit confirmation mode overrides the profile default
func TestEP_ExplicitConfirmationOverridesProfile(t *testing.T) {
	cfg := &config.Config{}
	cfg.Workflow.EnforcementProfile = config.ProfileOutcome
	cfg.Workflow.RawStepConfirmationMode = config.ConfirmEveryStep
	if !epRenders(&workflow.Workflow{CurrentStep: "code"}, cfg) {
		t.Fatal("explicit every_step must win over outcome and render the prompt")
	}
}
