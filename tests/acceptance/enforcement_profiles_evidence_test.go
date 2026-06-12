package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Acceptance: specs/enforcement-profiles.feature

// Scenario: Strict profile requires subagent orchestration evidence
func TestEP_StrictRequiresSubagentEvidence(t *testing.T) {
	wf := workflow.NewWithOrder("f", workflow.DefaultStepOrder, config.ProfileStrict)
	if wf.OrchestrationMode != workflow.StrictOrchestrationMode {
		t.Fatalf("strict must set %q, got %q", workflow.StrictOrchestrationMode, wf.OrchestrationMode)
	}
}

// Scenario: Guided profile does not require subagent orchestration evidence
func TestEP_GuidedNoSubagentEvidence(t *testing.T) {
	wf := workflow.NewWithOrder("f", workflow.DefaultStepOrder, config.ProfileGuided)
	if wf.OrchestrationMode != "" {
		t.Fatalf("guided must leave orchestration mode empty, got %q", wf.OrchestrationMode)
	}
}

// Scenario: Outcome profile does not require subagent orchestration evidence
func TestEP_OutcomeNoSubagentEvidence(t *testing.T) {
	wf := workflow.NewWithOrder("f", workflow.DefaultStepOrder, config.ProfileOutcome)
	if wf.OrchestrationMode != "" {
		t.Fatalf("outcome must leave orchestration mode empty, got %q", wf.OrchestrationMode)
	}
}

// Scenario: A per-feature profile overrides the global setting
func TestEP_PerFeatureOverridesGlobal(t *testing.T) {
	cfg := &config.Config{}
	cfg.Workflow.EnforcementProfile = config.ProfileGuided
	wf := &workflow.Workflow{EnforcementProfile: config.ProfileOutcome}
	if got := workflow.EffectiveProfile(wf, cfg); got != config.ProfileOutcome {
		t.Fatalf("per-feature override must win, got %q", got)
	}
}
