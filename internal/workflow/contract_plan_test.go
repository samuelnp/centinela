package workflow

import (
	"os"
	"testing"
)

// chdirTemp moves into a scratch repo root so Load() resolves .workflow/ there.
func chdirTemp(t *testing.T) {
	t.Helper()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := os.MkdirAll(WorkflowDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
}

func TestUsesUnifiedPlanner_PinnedContract(t *testing.T) {
	wf := &Workflow{Feature: "f", PlanContract: PlanContractUnified}
	if !wf.UsesUnifiedPlanner() {
		t.Fatal("pinned planner-v1 workflow should use the unified planner")
	}
}

func TestUsesUnifiedPlanner_EmptyContractIsLegacy(t *testing.T) {
	if (&Workflow{Feature: "f"}).UsesUnifiedPlanner() {
		t.Fatal("unpinned workflow must be treated as legacy")
	}
}

func TestUsesUnifiedPlanner_NilWorkflowIsLegacy(t *testing.T) {
	var wf *Workflow
	if wf.UsesUnifiedPlanner() {
		t.Fatal("nil workflow must be treated as legacy")
	}
}

func TestUsesUnifiedPlanner_UnknownContractIsLegacy(t *testing.T) {
	if (&Workflow{PlanContract: "planner-v0"}).UsesUnifiedPlanner() {
		t.Fatal("an unrecognized contract must not satisfy planner-v1")
	}
}

func TestFeatureUsesUnifiedPlanner_MissingStateFailsLegacy(t *testing.T) {
	chdirTemp(t)
	if FeatureUsesUnifiedPlanner("ghost") {
		t.Fatal("a missing state file must never invent the new gate")
	}
}

func TestFeatureUsesUnifiedPlanner_UnreadableStateFailsLegacy(t *testing.T) {
	chdirTemp(t)
	if err := os.WriteFile(FilePath("broken"), []byte("{not json"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if FeatureUsesUnifiedPlanner("broken") {
		t.Fatal("an unparseable state file must fail legacy")
	}
}

func TestFeatureUsesUnifiedPlanner_ReadsPinFromDisk(t *testing.T) {
	chdirTemp(t)
	if err := Save(New("pinned")); err != nil {
		t.Fatalf("save: %v", err)
	}
	if !FeatureUsesUnifiedPlanner("pinned") {
		t.Fatal("a freshly started workflow must resolve as planner-v1")
	}
	legacy := New("legacy")
	legacy.PlanContract = ""
	if err := Save(legacy); err != nil {
		t.Fatalf("save: %v", err)
	}
	if FeatureUsesUnifiedPlanner("legacy") {
		t.Fatal("a state file without planContract must resolve as legacy")
	}
}

func TestNewWithOrder_PinsPlanContract(t *testing.T) {
	wf := NewWithOrder("f", DefaultStepOrder, "strict")
	if wf.PlanContract != PlanContractUnified {
		t.Fatalf("PlanContract = %q, want %q", wf.PlanContract, PlanContractUnified)
	}
}
