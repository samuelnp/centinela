package main

import (
	"os"

	"github.com/samuelnp/centinela/internal/workflow"
	"github.com/samuelnp/centinela/internal/worktree"
)

func loadActiveWorkflows() []*workflow.Workflow {
	// Classification (evidence-leak rejection, dedupe, recency sort) lives in
	// the domain layer. cmd/ keeps ONLY the worktree-scoping filter: when cwd
	// is inside a worktree, that worktree's feature is authoritative.
	cwd, _ := os.Getwd()
	feature, _ := worktree.DetectFeatureFromCwd(cwd)
	active := workflow.ActiveWorkflows(workflow.WorkflowDir)
	if feature == "" {
		return active
	}
	out := active[:0]
	for _, wf := range active {
		if wf.Feature == feature {
			out = append(out, wf)
		}
	}
	return out
}
