package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRenderZeroCoverageFunctions(t *testing.T) {
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
	wf := &workflow.Workflow{CurrentStep: "done", Steps: map[string]workflow.StepState{"plan": {Status: "done"}}}
	if wfDoneCount(wf) != 4 || roadmapIcon("planned") == "" || stepStatusLine(wf, "plan", wf.Steps["plan"]) == "" {
		t.Fatal("expected helper outputs")
	}
}
