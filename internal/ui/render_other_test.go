package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRenderGatesAndStatusHelpers(t *testing.T) {
	if RenderGateResult(gates.Result{Name: "g", Status: gates.Pass, Message: "ok"}) == "" {
		t.Fatal("pass gate render should not be empty")
	}
	if !strings.Contains(RenderCmdResult("go test", false, "bad"), "bad") {
		t.Fatal("failed command output should include stderr")
	}
	wf := &workflow.Workflow{Feature: "f", CurrentStep: "code", Steps: map[string]workflow.StepState{"code": {Status: "in-progress"}}}
	if !strings.Contains(RenderStatus(wf), "Feature") || RenderSuccess("ok") == "" || RenderStep("Next", "tests") == "" {
		t.Fatal("status helpers should render")
	}
	if stepStatusLine(wf, "plan", workflow.StepState{Status: "pending"}) == "" {
		t.Fatal("stepStatusLine should render pending")
	}
}

func TestRenderRoadmapAndReview(t *testing.T) {
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "P1", Features: []roadmap.Feature{{Name: "new"}}}}}
	if !strings.Contains(RenderRoadmapNeeded(), "ROADMAP.md") {
		t.Fatal("roadmap needed output missing marker")
	}
	if !strings.Contains(RenderRoadmapSummary(r), "Roadmap") {
		t.Fatal("roadmap summary output missing label")
	}
	if !strings.Contains(RenderRoadmap(r), "P1") || roadmapIcon("done") == "" {
		t.Fatal("roadmap render should include phase and icon")
	}
	if !strings.Contains(RenderReviewReady("f", "plan", "code"), "shall I advance") {
		t.Fatal("review reminder missing expected prompt")
	}
}
