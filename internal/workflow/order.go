package workflow

import "time"

var DefaultStepOrder = []string{"plan", "code", "tests", "validate"}

var BootstrapStepOrder = []string{"plan", "code", "validate"}

var StepOrder = DefaultStepOrder

func NewWithOrder(feature string, order []string) *Workflow {
	steps := make(map[string]StepState, len(order))
	for i, step := range order {
		status := "pending"
		if i == 0 {
			status = "in-progress"
		}
		steps[step] = StepState{Status: status}
	}
	return &Workflow{
		Feature:     feature,
		StartedAt:   time.Now().UTC(),
		CurrentStep: order[0],
		Steps:       steps,
		StepOrder:   cloneOrder(order),
	}
}

func (wf *Workflow) OrderedSteps() []string {
	if wf == nil || len(wf.StepOrder) == 0 {
		return DefaultStepOrder
	}
	return wf.StepOrder
}

func cloneOrder(order []string) []string {
	out := make([]string, len(order))
	copy(out, order)
	return out
}
