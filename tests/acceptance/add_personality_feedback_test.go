package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Acceptance: specs/add-personality-feedback.feature
func TestPersonaOutputKeepsActionableContent(t *testing.T) {
	wf := workflow.New("f")
	wf.CurrentStep = "code"
	wf.Steps["plan"] = workflow.StepState{Status: "done"}
	wf.Steps["code"] = workflow.StepState{Status: "in-progress"}

	tag := ui.RenderTag(wf)
	blocked := ui.RenderBlocked("code", "plan", "f", "/tmp/a.go")

	if !strings.Contains(tag, "🛡️👁️") || !strings.Contains(tag, "HOOK") {
		t.Fatal("tag output should keep source metadata with persona")
	}
	if !strings.Contains(blocked, "Next action") || !strings.Contains(blocked, "BLOCKED WRITE") {
		t.Fatal("blocked output should keep reason and action hint")
	}
}
