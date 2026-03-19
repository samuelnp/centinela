package integration_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/hookpolicy"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestEvaluatePrewrite_AnyWorkflowCanAllow(t *testing.T) {
	cfg := &config.Config{}
	wfs := []*workflow.Workflow{
		{Feature: "a", CurrentStep: "plan"},
		{Feature: "b", CurrentStep: "code"},
	}
	d := hookpolicy.EvaluatePrewrite("/repo/internal/x.go", "/repo", cfg, wfs)
	if !d.Allow {
		t.Fatalf("expected allow when one workflow permits write, got %+v", d)
	}
}

func TestEvaluatePrewrite_IgnoresOutsideWorkspace(t *testing.T) {
	cfg := &config.Config{}
	wfs := []*workflow.Workflow{{Feature: "a", CurrentStep: "plan"}}
	d := hookpolicy.EvaluatePrewrite("/other/internal/x.go", "/repo", cfg, wfs)
	if !d.Allow {
		t.Fatalf("expected outside path to be ignored, got %+v", d)
	}
}
