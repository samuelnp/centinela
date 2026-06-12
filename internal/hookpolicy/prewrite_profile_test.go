package hookpolicy

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

func wfAt(step, profile string) []*workflow.Workflow {
	return []*workflow.Workflow{{Feature: "f", CurrentStep: step, EnforcementProfile: profile}}
}

// outcome drops the ordering rails: a code-typed write during the plan step is
// allowed (it would be blocked under strict/guided).
func TestEvaluatePrewrite_OutcomeBypassesStepGating(t *testing.T) {
	cfg := &config.Config{}
	d := EvaluatePrewrite("/repo/internal/a.go", "/repo", cfg, wfAt("plan", config.ProfileOutcome))
	if !d.Allow {
		t.Fatalf("outcome must allow code write during plan, got %+v", d)
	}
}

// strict AND guided keep today's step-gating: a code write in plan is blocked
// with workflow context.
func TestEvaluatePrewrite_StrictAndGuidedBlockOutOfStep(t *testing.T) {
	cfg := &config.Config{}
	for _, p := range []string{config.ProfileStrict, config.ProfileGuided} {
		d := EvaluatePrewrite("/repo/internal/a.go", "/repo", cfg, wfAt("plan", p))
		if d.Allow || d.Feature != "f" || d.Step != "plan" {
			t.Fatalf("profile %q must block code write in plan, got %+v", p, d)
		}
	}
}

// No active workflow blocks a plan/code write regardless of the global profile:
// outcome relaxes ordering, never the requirement that work happen under a
// started feature.
func TestEvaluatePrewrite_NoActiveWorkflowBlocksUnderAnyProfile(t *testing.T) {
	for _, p := range []string{config.ProfileStrict, config.ProfileGuided, config.ProfileOutcome} {
		cfg := cfgGlobal(p)
		d := EvaluatePrewrite("/repo/internal/a.go", "/repo", cfg, nil)
		if d.Allow || !d.NeedInit {
			t.Fatalf("global profile %q with no workflow must NeedInit-block, got %+v", p, d)
		}
		// A done-only workflow is likewise inactive → block.
		done := []*workflow.Workflow{{Feature: "f", CurrentStep: "done"}}
		if d2 := EvaluatePrewrite("/repo/internal/a.go", "/repo", cfg, done); d2.Allow || !d2.NeedInit {
			t.Fatalf("profile %q done-only must NeedInit-block, got %+v", p, d2)
		}
	}
}

func cfgGlobal(p string) *config.Config {
	c := &config.Config{}
	c.Workflow.EnforcementProfile = p
	return c
}
