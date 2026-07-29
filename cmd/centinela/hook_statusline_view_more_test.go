package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestPrimaryWorkflowSkipsRoleAndEmptyEntries(t *testing.T) {
	bad := &workflow.Workflow{}
	role := workflow.New("alpha-big-thinker")
	main := workflow.New("alpha")
	wf := primaryWorkflow([]*workflow.Workflow{bad, role, main})
	if wf == nil || wf.Feature != "alpha" {
		t.Fatalf("expected alpha workflow, got %#v", wf)
	}
}

func TestDoneCountDoneState(t *testing.T) {
	wf := workflow.New("alpha")
	wf.CurrentStep = "done"
	if got := doneCount(wf); got != len(wf.OrderedSteps()) {
		t.Fatalf("expected full done count, got %d", got)
	}
}

func TestPrimaryWorkflowFallsBackToSecondPass(t *testing.T) {
	role := workflow.New("alpha-big-thinker")
	wf := primaryWorkflow([]*workflow.Workflow{role})
	if wf == nil || wf.Feature != "alpha-big-thinker" {
		t.Fatalf("expected fallback workflow, got %#v", wf)
	}
}

// isRoleWorkflow gained the "-planner" suffix for the unified planner role;
// a planner role sub-workflow must never be picked as primary over the
// feature it was delegated for.
func TestPrimaryWorkflowSkipsPlannerRoleWorkflow(t *testing.T) {
	role := workflow.New("alpha-planner")
	main := workflow.New("alpha")
	wf := primaryWorkflow([]*workflow.Workflow{role, main})
	if wf == nil || wf.Feature != "alpha" {
		t.Fatalf("expected alpha workflow, got %#v", wf)
	}
}
