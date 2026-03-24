package workflow

import "testing"

func TestNewWithOrderAndOrderedSteps(t *testing.T) {
	wf := NewWithOrder("f", BootstrapStepOrder)
	steps := wf.OrderedSteps()
	if len(steps) != 3 || steps[2] != "validate" {
		t.Fatalf("unexpected step order: %#v", steps)
	}
	if StepNumberFor(wf, "validate") != 3 {
		t.Fatal("validate should be third step in bootstrap order")
	}
}

func TestStepIndexCompatibility(t *testing.T) {
	if stepIndex("validate") != 3 {
		t.Fatal("compatibility step index should use default order")
	}
}
