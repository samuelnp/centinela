package workflow

import (
	"time"

	"github.com/samuelnp/centinela/internal/config"
)

var DefaultStepOrder = []string{"plan", "code", "tests", "validate", "docs"}

var BootstrapStepOrder = []string{"plan", "code", "validate", "docs"}

const StrictOrchestrationMode = "strict-subagents-v1"

var StepOrder = DefaultStepOrder

// NewWithOrder builds a workflow under the given enforcement profile. The
// profile is pinned on the workflow and decides orchestration evidence: only
// strict (RequireSubagentEvidence) sets StrictOrchestrationMode, so guided and
// outcome leave it empty and validateOrchestration early-returns for them.
func NewWithOrder(feature string, order []string, profile string) *Workflow {
	profile = config.NormalizeEnforcementProfile(profile)
	steps := make(map[string]StepState, len(order))
	for i, step := range order {
		status := "pending"
		if i == 0 {
			status = "in-progress"
		}
		steps[step] = StepState{Status: status}
	}
	mode := ""
	if config.ProfileDefaults(profile).RequireSubagentEvidence {
		mode = StrictOrchestrationMode
	}
	return &Workflow{
		Feature:            feature,
		StartedAt:          time.Now().UTC(),
		CurrentStep:        order[0],
		Steps:              steps,
		StepOrder:          cloneOrder(order),
		OrchestrationMode:  mode,
		EnforcementProfile: profile,
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
