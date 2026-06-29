package main

import (
	"github.com/samuelnp/centinela/internal/workflow"
	"github.com/samuelnp/centinela/internal/worktree"
)

// activeWorkflow resolves the active feature's workflow state for cost
// attribution: the worktree feature when cwd is inside a `.worktrees/<feature>`,
// otherwise the most-recently-touched active workflow in the root .workflow dir
// (non-worktree mode). Returns nil when no feature is active.
func activeWorkflow(cwd string) *workflow.Workflow {
	if feature, _ := worktree.DetectFeatureFromCwd(cwd); feature != "" {
		if wf, err := workflow.Load(feature); err == nil {
			return wf
		}
	}
	if wfs := workflow.ActiveWorkflows(workflow.WorkflowDir); len(wfs) > 0 {
		return wfs[0]
	}
	return nil
}
