package hookpolicy

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestEvaluatePrewriteBranches(t *testing.T) {
	cfg := &config.Config{}
	if !EvaluatePrewrite("/x/internal/a.go", "/repo", cfg, nil).Allow {
		t.Fatal("outside workspace should be allowed")
	}
	if !EvaluatePrewrite("/repo/docs/features/f.md", "/repo", cfg, nil).Allow {
		t.Fatal("roadmap writes should be allowed")
	}
	wfs := []*workflow.Workflow{{Feature: "f", CurrentStep: "code"}}
	if !EvaluatePrewrite("/repo/internal/a.go", "/repo", cfg, wfs).Allow {
		t.Fatal("code step should allow code files")
	}
	wfs2 := []*workflow.Workflow{{Feature: "f", CurrentStep: "plan"}}
	d := EvaluatePrewrite("/repo/internal/a.go", "/repo", cfg, wfs2)
	if d.Allow || d.Feature != "f" || d.Step != "plan" {
		t.Fatalf("expected block with context, got %+v", d)
	}
	d2 := EvaluatePrewrite("/repo/internal/a.go", "/repo", cfg, []*workflow.Workflow{{Feature: "f", CurrentStep: "done"}})
	if !d2.NeedInit || d2.Allow {
		t.Fatalf("done workflow should require new start, got %+v", d2)
	}
}

func TestIsInsideWorkspace(t *testing.T) {
	if !isInsideWorkspace("/a/b", "") {
		t.Fatal("empty cwd should allow")
	}
	if isInsideWorkspace("rel/path", "/abs") {
		t.Fatal("mixed abs/rel should not be inside workspace")
	}
}
