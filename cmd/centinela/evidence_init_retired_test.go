package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// writeLegacyWorkflow saves a workflow that predates planner-v1.
func writeLegacyWorkflow(t *testing.T, feature string) {
	t.Helper()
	wf := workflow.New(feature)
	wf.PlanContract = ""
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
}

// D7: a pinned workflow refuses the retired role and points at planner.
func TestEvidenceInitRefusesRetiredRoleOnPinnedWorkflow(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "fresh")
	err := runEvidenceInit(nil, []string{"fresh", "big-thinker"})
	if err == nil {
		t.Fatal("expected a pinned workflow to refuse the retired role")
	}
	if !strings.Contains(err.Error(), "retired") || !strings.Contains(err.Error(), "planner") {
		t.Fatalf("error must be retired + planner pointer, got %v", err)
	}
	if _, statErr := os.Stat(".workflow/fresh-big-thinker.json"); !os.IsNotExist(statErr) {
		t.Fatal("refused init must not write a stub")
	}
}

// D7 back-compat: an in-flight legacy workflow can still author its pair.
func TestEvidenceInitAllowsRetiredRoleOnLegacyWorkflow(t *testing.T) {
	chdirEvidenceTemp(t)
	writeLegacyWorkflow(t, "old")
	if err := runEvidenceInit(nil, []string{"old", "big-thinker"}); err != nil {
		t.Fatalf("legacy workflow must still author big-thinker: %v", err)
	}
	if _, err := os.Stat(".workflow/old-big-thinker.json"); err != nil {
		t.Fatalf("legacy stub not written: %v", err)
	}
	if err := runEvidenceInit(nil, []string{"old", "feature-specialist"}); err != nil {
		t.Fatalf("legacy workflow must still author feature-specialist: %v", err)
	}
}

// D7 no-workflow case: a retired role against a feature with NO workflow
// state at all must still say "retired" + point at planner, not the generic
// "unknown feature" message — EnsureRoleAllowed runs before requireKnownFeature.
func TestEvidenceInitRefusesRetiredRoleWithNoWorkflowAtAll(t *testing.T) {
	chdirEvidenceTemp(t)
	err := runEvidenceInit(nil, []string{"ghost-feature", "feature-specialist"})
	if err == nil {
		t.Fatal("expected an error for a retired role with no workflow state")
	}
	if !strings.Contains(err.Error(), "retired") || !strings.Contains(err.Error(), "planner") {
		t.Fatalf("error must be retired + planner pointer, got %v", err)
	}
	if strings.Contains(err.Error(), "unknown feature") {
		t.Fatalf("must not fall back to the generic unknown-feature message: %v", err)
	}
}

// The planner stub itself is always authorable and lands at the planner path.
func TestEvidenceInitWritesPlannerStub(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "fresh")
	if err := runEvidenceInit(nil, []string{"fresh", "planner"}); err != nil {
		t.Fatalf("planner init: %v", err)
	}
	data, err := os.ReadFile(".workflow/fresh-planner.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"role": "planner"`, `"step": "plan"`, `"handoffTo": "senior-engineer"`} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("planner stub missing %s: %s", want, data)
		}
	}
}
