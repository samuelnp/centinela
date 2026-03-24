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

func TestOrderedStepsFallbackAndStepNumberForNil(t *testing.T) {
	wf := &Workflow{}
	if len(wf.OrderedSteps()) != 4 {
		t.Fatal("empty workflow should fallback to default step order")
	}
	if StepNumberFor(nil, "tests") != 3 {
		t.Fatal("nil workflow should fallback to default step numbering")
	}
}
