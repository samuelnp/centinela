package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRenderRoadmapSummaryNoInProgress(t *testing.T) {
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Features: []roadmap.Feature{{Name: "missing"}}}}}
	line := RenderRoadmapSummary(r)
	if !strings.Contains(line, "Roadmap") {
		t.Fatal("summary should include label")
	}
}

func TestStepStatusLineCompletedWithDate(t *testing.T) {
	d := "2026-03-19T00:00:00Z"
	info := workflow.StepState{Status: "done", CompletedAt: &d}
	wf := &workflow.Workflow{CurrentStep: "tests", Steps: map[string]workflow.StepState{"plan": info}}
	out := stepStatusLine(wf, "plan", info)
	if !strings.Contains(out, "done") {
		t.Fatal("expected done status rendering")
	}
}
