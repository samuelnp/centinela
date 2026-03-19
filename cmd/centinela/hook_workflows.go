package main

import (
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/workflow"
)

func loadActiveWorkflows() []*workflow.Workflow {
	entries, _ := filepath.Glob(filepath.Join(workflow.WorkflowDir, "*.json"))
	var out []*workflow.Workflow
	for _, p := range entries {
		wf, err := workflow.Load(strings.TrimSuffix(filepath.Base(p), ".json"))
		if err == nil {
			out = append(out, wf)
		}
	}
	return out
}
