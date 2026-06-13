package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Headless via env short-circuits shouldRenderReviewPrompt to false even when
// the explicit knob is every_step — headless wins over every_step.
func TestShouldRenderReviewPrompt_HeadlessBeatsEveryStep(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "1")
	cfg := &config.Config{}
	cfg.Workflow.RawStepConfirmationMode = "every_step"
	wf := &workflow.Workflow{CurrentStep: "validate"}
	if shouldRenderReviewPrompt(wf, cfg) {
		t.Fatal("headless must suppress the review prompt under every_step")
	}
}

// Headless via [headless] config also suppresses the prompt under every_step.
func TestShouldRenderReviewPrompt_HeadlessConfigSuppresses(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	cfg := &config.Config{}
	cfg.Headless.Enabled = true
	cfg.Workflow.RawStepConfirmationMode = "every_step"
	wf := &workflow.Workflow{CurrentStep: "validate"}
	if shouldRenderReviewPrompt(wf, cfg) {
		t.Fatal("headless config must suppress the review prompt")
	}
}

// Back-compat: with headless off and every_step, the prompt still renders.
func TestShouldRenderReviewPrompt_HeadlessOffStillRenders(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	cfg := &config.Config{}
	cfg.Workflow.RawStepConfirmationMode = "every_step"
	wf := &workflow.Workflow{CurrentStep: "validate"}
	if !shouldRenderReviewPrompt(wf, cfg) {
		t.Fatal("headless off + every_step must still render the prompt")
	}
}
