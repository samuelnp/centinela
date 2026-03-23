package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRenderZeroCoverageFunctions(t *testing.T) {
	if RenderSetupNeeded() == "" {
		t.Fatal("expected setup needed output")
	}
	if RenderCmdResult("go test", true, "") == "" {
		t.Fatal("expected passing cmd render output")
	}
	if !strings.Contains(RenderFeatureBriefNeeded("f"), "Feature brief") {
		t.Fatal("expected feature brief output")
	}
	if !strings.Contains(RenderProductionReadinessSetupNeeded(), "Production readiness") {
		t.Fatal("expected production readiness setup output")
	}
	if !strings.Contains(RenderProductionReadinessWarning("f"), "WARNING") {
		t.Fatal("expected warning output")
	}
	if RenderGateResult(gates.Result{Name: "g", Status: gates.Fail, Message: "m", Details: []string{"d"}}) == "" {
		t.Fatal("fail gate output should not be empty")
	}
	if RenderGateResult(gates.Result{Name: "g", Status: gates.Warn, Message: "m"}) == "" {
		t.Fatal("warn gate output should not be empty")
	}
	if RenderGateResult(gates.Result{Name: "g", Status: gates.Skip, Message: "m"}) == "" {
		t.Fatal("skip gate output should not be empty")
	}
	if RenderGateResult(gates.Result{Name: "g", Status: 999, Message: "m"}) != "" {
		t.Fatal("unknown gate status should render empty string")
	}
	wf := &workflow.Workflow{CurrentStep: "done", Steps: map[string]workflow.StepState{"plan": {Status: "done"}}}
	wf2 := &workflow.Workflow{CurrentStep: "code", Steps: map[string]workflow.StepState{"code": {Status: "in-progress"}}}
	wf3 := &workflow.Workflow{CurrentStep: "unknown", Steps: map[string]workflow.StepState{}}
	if wfDoneCount(wf) != 4 || wfDoneCount(wf3) != 0 || roadmapIcon("planned") == "" || roadmapIcon("in-progress") == "" || stepStatusLine(wf2, "code", wf2.Steps["code"]) == "" || stepStatusLine(wf, "plan", wf.Steps["plan"]) == "" {
		t.Fatal("expected helper outputs")
	}
}
