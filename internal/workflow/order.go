package workflow

import (
	"time"

	"github.com/samuelnp/centinela/internal/config"
)

// DefaultStepOrder is the canonical five-step track.
var DefaultStepOrder = []string{"plan", "code", "tests", "validate", "docs"}

// BootstrapStepOrder is the four-step track for Phase 0 bootstrap features,
// which have no separate tests step.
var BootstrapStepOrder = []string{"plan", "code", "validate", "docs"}

// StrictOrchestrationMode is the orchestration-evidence mode pinned at start
// under the strict profile; it is what makes the subagent evidence bundle
// required. Guided and outcome leave the field empty.
const StrictOrchestrationMode = "strict-subagents-v1"

// StepOrder is the package-level default order used by callers that predate
// per-workflow step orders.
var StepOrder = DefaultStepOrder

// New creates a fresh workflow starting at the "plan" step under the strict
// (back-compat default) profile.
func New(feature string) *Workflow {
	return NewWithOrder(feature, DefaultStepOrder, config.ProfileStrict)
}

// NewWithOrder builds a workflow under the given enforcement profile. The
// profile is pinned on the workflow and decides orchestration evidence: only
// strict (RequireSubagentEvidence) sets StrictOrchestrationMode, so guided and
// outcome leave it empty and validateOrchestration early-returns for them.
// Every workflow born here also carries ProfileContractGuidedDefault, which is
// what entitles it to the guided tail default (see profile_contract.go).
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
		ValidateContract:   ValidateContractAdversarial,
		PlanContract:       PlanContractUnified,
		ProfileContract:    ProfileContractGuidedDefault,
	}
}

// OrderedSteps returns this workflow's pinned step order, falling back to
// DefaultStepOrder for a nil or unpinned workflow.
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
