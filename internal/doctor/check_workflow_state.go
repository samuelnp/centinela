package doctor

import (
	"path/filepath"
	"sort"

	"github.com/samuelnp/centinela/internal/workflow"
	"github.com/samuelnp/centinela/internal/worktree"
)

// workflowStateCheck reports per-feature workflow state files under .workflow/
// whose feature has neither a live branch nor an active worktree (orphaned).
// REPORT-ONLY: --fix never deletes workflow state; it surfaces the exact `rm`
// command. Per-role evidence JSONs and roadmap*.json are not feature state.
type workflowStateCheck struct{}

func (workflowStateCheck) Name() string { return "workflow-state" }

func (workflowStateCheck) Run(ctx Context) Diagnosis {
	d := Diagnosis{Name: "workflow-state"}
	orphans := orphanedWorkflows(ctx.Root)
	if len(orphans) == 0 {
		d.Status = OK
		d.Message = "no orphaned .workflow state"
		return d
	}
	d.Status = Error
	d.Message = "orphaned .workflow state — delete manually"
	var cmds []string
	for _, o := range orphans {
		path := filepath.Join(".workflow", o+".json")
		cmd := "rm " + path
		d.Details = append(d.Details, o+": "+cmd)
		cmds = append(cmds, cmd)
	}
	d.Repair = &Repair{Command: joinCommands(cmds)}
	return d
}

// orphanedWorkflows returns sorted feature names with a workflow state file but
// no live branch and no worktree directory.
func orphanedWorkflows(root string) []string {
	wfs := workflow.ActiveWorkflows(filepath.Join(root, ".workflow"))
	gitOK := gitAvailable(root)
	var out []string
	for _, wf := range wfs {
		if worktree.Exists(root, wf.Feature) {
			continue
		}
		if gitOK && branchExists(root, wf.Feature) {
			continue
		}
		out = append(out, wf.Feature)
	}
	sort.Strings(out)
	return out
}

// branchExists reports whether refs/heads/<feature> exists.
func branchExists(root, feature string) bool {
	_, err := gitRunner(root, "rev-parse", "--verify", "refs/heads/"+feature)
	return err == nil
}
