package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/hookpolicy"
	"github.com/samuelnp/centinela/internal/workflow"
)

func epWorkflows(step, profile string) []*workflow.Workflow {
	return []*workflow.Workflow{{Feature: "f", CurrentStep: step, EnforcementProfile: profile}}
}

// Acceptance: specs/enforcement-profiles.feature

// Scenario: Outcome profile allows writing code during the plan step
func TestEP_OutcomeAllowsCodeWriteDuringPlan(t *testing.T) {
	d := hookpolicy.EvaluatePrewrite("/repo/internal/a.go", "/repo", &config.Config{},
		epWorkflows("plan", config.ProfileOutcome))
	if !d.Allow {
		t.Fatalf("outcome must allow code write in plan step, got %+v", d)
	}
}

// Scenario: Strict and guided profiles still block out-of-step writes
func TestEP_StrictAndGuidedBlockOutOfStepWrites(t *testing.T) {
	for _, p := range []string{config.ProfileStrict, config.ProfileGuided} {
		d := hookpolicy.EvaluatePrewrite("/repo/internal/a.go", "/repo", &config.Config{},
			epWorkflows("plan", p))
		if d.Allow {
			t.Fatalf("profile %q must block code write in plan step", p)
		}
	}
}

// Scenario: A write with no active workflow is always blocked
func TestEP_NoActiveWorkflowAlwaysBlocked(t *testing.T) {
	for _, p := range []string{config.ProfileStrict, config.ProfileGuided, config.ProfileOutcome} {
		cfg := &config.Config{}
		cfg.Workflow.EnforcementProfile = p
		d := hookpolicy.EvaluatePrewrite("/repo/internal/a.go", "/repo", cfg, nil)
		if d.Allow || !d.NeedInit {
			t.Fatalf("global profile %q with no workflow must block, got %+v", p, d)
		}
	}
}
