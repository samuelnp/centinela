package workflow

import (
	"fmt"
	"time"

	"github.com/samuelnp/centinela/internal/config"
)

// Complete marks the current step done and advances to the next one.
func (wf *Workflow) Complete(cfg *config.Config) error {
	current := wf.CurrentStep
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

	nextIdx := stepIndex(current) + 1
	if nextIdx >= len(StepOrder) {
		wf.CurrentStep = "done"
		return nil
	}

	next := StepOrder[nextIdx]
	wf.CurrentStep = next
	ns := wf.Steps[next]
	ns.Status = "in-progress"
	wf.Steps[next] = ns
	return nil
}

// StepNumber returns the 1-based position of a step (1=plan … 4=validate).
func StepNumber(step string) int {
	return stepIndex(step) + 1
}

// StepIndex returns the zero-based index of a step name, or -1 if not found.
func stepIndex(step string) int {
	for i, s := range StepOrder {
		if s == step {
			return i
		}
	}
	return -1
}
