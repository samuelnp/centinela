package hookpolicy

import (
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/workflow"
)

// firstDegraded returns the first active unmodellable workflow, so a stale-binary
// refusal can name the feature whose state file this binary cannot read.
func firstDegraded(wfs []*workflow.Workflow) *workflow.Workflow {
	for _, wf := range wfs {
		if wf.CurrentStep != "done" && wf.Unmodellable() {
			return wf
		}
	}
	return &workflow.Workflow{}
}

// firstEnforceable returns the first active workflow this binary can model, so
// a refusal names a step the author can actually act on rather than a degraded
// file's unknown one.
func firstEnforceable(wfs []*workflow.Workflow) *workflow.Workflow {
	for _, wf := range wfs {
		if wf.CurrentStep != "done" && !wf.Unmodellable() {
			return wf
		}
	}
	return &workflow.Workflow{}
}

func isInsideWorkspace(path, cwd string) bool {
	if cwd == "" {
		return true
	}
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}
