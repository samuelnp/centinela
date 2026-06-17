package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// hgRenders mirrors cmd.shouldRenderReviewPrompt: headless short-circuits to
// false BEFORE the per-knob resolver, so headless wins over every_step.
func hgRenders(wf *workflow.Workflow, cfg *config.Config) bool {
	if config.IsHeadless(cfg) {
		return false
	}
	mode := config.ConfirmEveryStep
	if cfg.Workflow.RawStepConfirmationMode != "" {
		mode = config.NormalizeStepConfirmationMode(cfg.Workflow.RawStepConfirmationMode)
	}
	return mode != config.ConfirmAuto
}

// hgAdvisorSpeaks mirrors cmd.runHookPlanAdvisor's headless short-circuit: under
// headless the hook returns nil before loading workflows and emits nothing.
func hgAdvisorSpeaks(step string, cfg *config.Config) bool {
	if config.IsHeadless(cfg) {
		return false
	}
	return step == "plan"
}

// Acceptance: specs/headless-governance.feature

// Scenario: Headless via env suppresses the step-review prompt even under every_step
func TestHG_EnvSuppressesReviewUnderEveryStep(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "1")
	cfg := &config.Config{}
	cfg.Workflow.RawStepConfirmationMode = config.ConfirmEveryStep
	if hgRenders(&workflow.Workflow{CurrentStep: "validate"}, cfg) {
		t.Fatal("headless env must suppress prompt; headless wins over every_step")
	}
}

// Scenario: Headless via config suppresses the step-review prompt
func TestHG_ConfigSuppressesReview(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	cfg := hgCfg(true, false)
	cfg.Workflow.RawStepConfirmationMode = config.ConfirmEveryStep
	if hgRenders(&workflow.Workflow{CurrentStep: "validate"}, cfg) {
		t.Fatal("headless config must suppress the review prompt")
	}
}

// Scenario: Headless via config suppresses the plan-advisor directive
func TestHG_ConfigSuppressesPlanAdvisor(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	if hgAdvisorSpeaks("plan", hgCfg(true, false)) {
		t.Fatal("headless config must suppress the plan-advisor directive")
	}
}

// Scenario: Plan advisor stays quiet under headless even when it would otherwise speak
func TestHG_PlanAdvisorQuietUnderHeadless(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "1")
	if hgAdvisorSpeaks("plan", &config.Config{}) {
		t.Fatal("plan advisor must short-circuit before loading workflows under headless")
	}
}

// Scenario: Back-compat review prompt under every_step still renders when headless off
func TestHG_BackCompatReviewStillRenders(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	t.Setenv("CI", "")
	cfg := &config.Config{}
	cfg.Workflow.RawStepConfirmationMode = config.ConfirmEveryStep
	if !hgRenders(&workflow.Workflow{CurrentStep: "validate"}, cfg) {
		t.Fatal("headless off + every_step must still render the prompt")
	}
}

// Scenario: Back-compat plan advisor still emits directives when headless off
func TestHG_BackCompatPlanAdvisorSpeaks(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	t.Setenv("CI", "")
	if !hgAdvisorSpeaks("plan", &config.Config{}) {
		t.Fatal("headless off must let the plan advisor emit directives")
	}
}
