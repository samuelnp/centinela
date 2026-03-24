package integration_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRenderOutputsAreExplicitlySystemBranded(t *testing.T) {
	wf := workflow.New("f")
	wf.CurrentStep = "code"
	wf.Steps["plan"] = workflow.StepState{Status: "done"}
	wf.Steps["code"] = workflow.StepState{Status: "in-progress"}
	tag := ui.RenderTag(wf)
	ctx := ui.RenderContext([]*workflow.Workflow{wf})
	blocked := ui.RenderBlocked("code", "plan", "f", "/tmp/a.go")
	if !strings.Contains(tag, "CENTINELA") || !strings.Contains(tag, "HOOK") {
		t.Fatal("tag output should be system branded")
	}
	if !strings.Contains(ctx, "ACTIVE WORKFLOWS") || !strings.Contains(ctx, "CENTINELA") {
		t.Fatal("context output should be explicitly branded")
	}
	if !strings.Contains(blocked, "BLOCKED WRITE") || !strings.Contains(blocked, "CENTINELA") || !strings.Contains(blocked, "Next action") {
		t.Fatal("blocked output should show explicit system header")
	}
}
