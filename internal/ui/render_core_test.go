package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func sampleWorkflow(step string) *workflow.Workflow {
	return &workflow.Workflow{Feature: "f", CurrentStep: step, Steps: map[string]workflow.StepState{
		"plan": {Status: "done"}, "code": {Status: "in-progress"}, "tests": {Status: "pending"}, "validate": {Status: "pending"},
	}}
}

func TestRenderBlockedTagContext(t *testing.T) {
	b := RenderBlocked("code", "plan", "f", "/tmp/a.go")
	if !strings.Contains(b, "BLOCKED") || !strings.Contains(b, "a.go") || !strings.Contains(b, "Next action") {
		t.Fatalf("unexpected blocked output: %q", b)
	}
	tg := RenderTag(sampleWorkflow("code"))
	if !strings.Contains(tg, "f") || !strings.Contains(tg, "code") {
		t.Fatalf("unexpected tag: %q", tg)
	}
	c := RenderContext([]*workflow.Workflow{sampleWorkflow("code")})
	if !strings.Contains(c, "f") || !strings.Contains(c, "code") {
		t.Fatalf("unexpected context: %q", c)
	}
}

func TestStepHelpers(t *testing.T) {
	wf := sampleWorkflow("code")
	if wfDoneCount(wf) != 1 {
		t.Fatalf("expected done count 1, got %d", wfDoneCount(wf))
	}
	if !strings.Contains(stepBar(wf), "plan") {
		t.Fatal("step bar should include step names")
	}
	if stepIcon(wf, "code") == "" || stepIcon(wf, "plan") == "" {
		t.Fatal("expected non-empty icons")
	}
}
