package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/workflow"
	"github.com/samuelnp/centinela/internal/worktree"
)

func loadActiveWorkflows() []*workflow.Workflow {
	// When cwd is inside a worktree, only that worktree's .workflow/ is
	// authoritative. Otherwise fall back to the main checkout's .workflow/.
	cwd, _ := os.Getwd()
	feature, _ := worktree.DetectFeatureFromCwd(cwd)
	entries, _ := filepath.Glob(filepath.Join(workflow.WorkflowDir, "*.json"))
	var out []*workflow.Workflow
	for _, p := range entries {
		name := strings.TrimSuffix(filepath.Base(p), ".json")
		wf, err := workflow.Load(name)
		if err != nil || wf.CurrentStep == "done" {
			continue
		}
		if feature != "" && wf.Feature != feature {
			continue
		}
		out = append(out, wf)
	}
	return out
}
