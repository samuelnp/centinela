package workflow

import (
	"fmt"
	"time"

	"github.com/samuelnp/centinela/internal/config"
)

// Complete marks the current step done and advances to the next one.
func (wf *Workflow) Complete(cfg *config.Config) error {
	current := wf.CurrentStep
	order := wf.OrderedSteps()
	if current == "done" {
		return fmt.Errorf("workflow already complete")
	}
	if err := ValidateArtifacts(wf.Feature, current, cfg); err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	step := wf.Steps[current]
	step.Status = "done"
	step.CompletedAt = &now
	wf.Steps[current] = step

	nextIdx := stepIndexIn(current, order) + 1
	if nextIdx >= len(order) {
		wf.CurrentStep = "done"
		return nil
	}

	next := order[nextIdx]
	wf.CurrentStep = next
	ns := wf.Steps[next]
	ns.Status = "in-progress"
	wf.Steps[next] = ns
	return nil
}

// StepNumber returns the 1-based position of a step (1=plan … 4=validate).
func StepNumber(step string) int {
	return StepNumberIn(DefaultStepOrder, step)
}

func StepNumberFor(wf *Workflow, step string) int {
	if wf == nil {
		return StepNumber(step)
	}
	return StepNumberIn(wf.OrderedSteps(), step)
}

func StepNumberIn(order []string, step string) int {
	return stepIndexIn(step, order) + 1
}

// StepIndex returns the zero-based index of a step name, or -1 if not found.
func stepIndexIn(step string, order []string) int {
	for i, s := range order {
		if s == step {
			return i
		}
	}
	return -1
}

func stepIndex(step string) int {
	return stepIndexIn(step, DefaultStepOrder)
}
