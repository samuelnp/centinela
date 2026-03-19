package unit_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/hookpolicy"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestEvaluatePrewrite_NoWorkflowNeedsStart(t *testing.T) {
	cfg := &config.Config{}
	d := hookpolicy.EvaluatePrewrite("/repo/internal/service/x.go", "/repo", cfg, nil)
	if d.Allow || !d.NeedInit || d.FileType != workflow.TypeCode {
		t.Fatalf("unexpected decision: %+v", d)
	}
}

func TestEvaluatePrewrite_AllowsRoadmapAndOther(t *testing.T) {
	cfg := &config.Config{}
	for _, p := range []string{"/repo/docs/features/a.md", "/repo/README.md"} {
		d := hookpolicy.EvaluatePrewrite(p, "/repo", cfg, nil)
		if !d.Allow {
			t.Fatalf("expected allow for %s, got %+v", p, d)
		}
	}
}

func TestEvaluatePrewrite_BlockedCarriesWorkflowContext(t *testing.T) {
	cfg := &config.Config{}
	wfs := []*workflow.Workflow{{Feature: "f1", CurrentStep: "plan"}}
	d := hookpolicy.EvaluatePrewrite("/repo/internal/service/x.go", "/repo", cfg, wfs)
	if d.Allow || d.NeedInit {
		t.Fatalf("expected blocked decision, got %+v", d)
	}
	if d.Feature != "f1" || d.Step != "plan" {
		t.Fatalf("missing context fields: %+v", d)
	}
}
