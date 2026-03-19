package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/hookpolicy"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Acceptance: specs/refactor-hook-policy-core.feature
func TestHookPolicy_RoadmapAlwaysAllowed(t *testing.T) {
	cfg := &config.Config{}
	wfs := []*workflow.Workflow{{Feature: "f", CurrentStep: "plan"}}
	d := hookpolicy.EvaluatePrewrite("/repo/docs/features/x.md", "/repo", cfg, wfs)
	if !d.Allow {
		t.Fatalf("roadmap write should be allowed, got %+v", d)
	}
}

func TestHookPolicy_BlockIncludesFeatureAndStep(t *testing.T) {
	cfg := &config.Config{}
	wfs := []*workflow.Workflow{{Feature: "f", CurrentStep: "plan"}}
	d := hookpolicy.EvaluatePrewrite("/repo/internal/x.go", "/repo", cfg, wfs)
	if d.Allow || d.Feature != "f" || d.Step != "plan" {
		t.Fatalf("expected blocked decision with context, got %+v", d)
	}
}
