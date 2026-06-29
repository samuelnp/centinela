package integration_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Integration: the active-workflows panel (a real CLI/hook panel) renders
// border-free while keeping its branded header and body.
func TestActiveWorkflowsPanelHasNoBorder(t *testing.T) {
	wf := workflow.New("f")
	wf.CurrentStep = "code"
	wf.Steps["plan"] = workflow.StepState{Status: "done"}
	wf.Steps["code"] = workflow.StepState{Status: "in-progress"}

	out := ui.RenderContext([]*workflow.Workflow{wf})
	if strings.ContainsAny(out, "╭╮╰╯│") {
		t.Fatalf("active-workflows panel should have no border, got:\n%s", out)
	}
	for _, want := range []string{"🛡️👁️", "ACTIVE WORKFLOWS", "f", "code"} {
		if !strings.Contains(out, want) {
			t.Errorf("panel lost content %q", want)
		}
	}
}
