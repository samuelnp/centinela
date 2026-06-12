package roadmap

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// doneWF persists a workflow whose current step is "done" so FeatureStatus
// resolves the feature as completed.
func doneWF(t *testing.T, name string) {
	t.Helper()
	if err := workflow.Save(&workflow.Workflow{
		Feature: name, CurrentStep: "done", Steps: map[string]workflow.StepState{},
	}); err != nil {
		t.Fatalf("save workflow %q: %v", name, err)
	}
}

// TestSummaryExcludesBacklog: every schedulable feature done plus a non-empty
// Backlog phase must report complete (planned/inProgress both zero), so the
// session-rehydration "Roadmap complete" path can fire.
func TestSummaryExcludesBacklog(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                       //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck

	doneWF(t, "f1")
	doneWF(t, "f2")
	r := &Roadmap{Phases: []Phase{
		{Name: "Phase 1", Features: []Feature{{Name: "f1"}, {Name: "f2"}}},
		{Name: BacklogPhaseName, Features: []Feature{{Name: "deferred-finding"}}},
	}}

	planned, inProgress, done := r.Summary()
	if planned != 0 || inProgress != 0 {
		t.Fatalf("backlog must not count: planned=%d inProgress=%d", planned, inProgress)
	}
	if done != 2 {
		t.Fatalf("done should count only schedulable features, got %d", done)
	}
	if planned > 0 || inProgress > 0 {
		t.Fatal("hasIncomplete must be false when all real features are done")
	}
}

// TestSummaryIncompleteWithBacklog: an incomplete schedulable feature still
// reports incomplete even alongside Backlog entries (Backlog never masks it).
func TestSummaryIncompleteWithBacklog(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                       //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck

	doneWF(t, "f1")
	r := &Roadmap{Phases: []Phase{
		{Name: "Phase 1", Features: []Feature{{Name: "f1"}, {Name: "f2"}}},
		{Name: BacklogPhaseName, Features: []Feature{{Name: "deferred-a"}, {Name: "deferred-b"}}},
	}}

	planned, inProgress, done := r.Summary()
	if planned != 1 {
		t.Fatalf("incomplete schedulable feature must count as planned, got %d", planned)
	}
	if inProgress != 0 || done != 1 {
		t.Fatalf("unexpected counts: inProgress=%d done=%d", inProgress, done)
	}
	if !(planned > 0 || inProgress > 0) {
		t.Fatal("hasIncomplete must be true with an incomplete schedulable feature")
	}
}
