package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

func cfgRaw(raw, profile string) *config.Config {
	c := &config.Config{}
	c.Workflow.RawStepConfirmationMode = raw
	c.Workflow.EnforcementProfile = profile
	return c
}

// outcome (no explicit knob) resolves to auto, suppressing the review prompt at
// every step.
func TestShouldRenderReviewPrompt_OutcomeSuppresses(t *testing.T) {
	cfg := cfgRaw("", config.ProfileOutcome)
	for _, step := range []string{"plan", "code", "tests", "validate", "docs"} {
		wf := &workflow.Workflow{CurrentStep: step}
		if shouldRenderReviewPrompt(wf, cfg) {
			t.Fatalf("outcome must suppress prompt at step %q", step)
		}
	}
}

// An explicit every_step knob overrides the outcome profile default: the prompt
// is rendered even though outcome alone would suppress it.
func TestShouldRenderReviewPrompt_ExplicitEveryStepBeatsOutcome(t *testing.T) {
	cfg := cfgRaw("every_step", config.ProfileOutcome)
	wf := &workflow.Workflow{CurrentStep: "code"}
	if !shouldRenderReviewPrompt(wf, cfg) {
		t.Fatal("explicit every_step must override outcome and render the prompt")
	}
}

func TestEffectiveConfirmationMode_Precedence(t *testing.T) {
	// strict → every_step; guided → after_plan; outcome → auto (all via profile).
	wf := &workflow.Workflow{CurrentStep: "code"}
	cases := map[string]string{
		config.ProfileStrict:  config.ConfirmEveryStep,
		config.ProfileGuided:  config.ConfirmAfterPlan,
		config.ProfileOutcome: config.ConfirmAuto,
	}
	for profile, want := range cases {
		if got := effectiveConfirmationMode(wf, cfgRaw("", profile)); got != want {
			t.Fatalf("profile %q: mode = %q, want %q", profile, got, want)
		}
	}
	// Explicit raw knob beats the profile default.
	if got := effectiveConfirmationMode(wf, cfgRaw("after_plan", config.ProfileStrict)); got != config.ConfirmAfterPlan {
		t.Fatalf("explicit after_plan must win over strict, got %q", got)
	}
}
