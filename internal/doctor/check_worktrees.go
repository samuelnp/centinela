package doctor

import (
	"path/filepath"
	"sort"

	"github.com/samuelnp/centinela/internal/worktree"
)

// worktreesCheck reports worktrees under .worktrees/ whose feature workflow is
// complete/absent or whose branch is merged into main. It is REPORT-ONLY:
// --fix never removes a worktree — it surfaces the exact `git worktree remove`
// command the operator must run themselves.
type worktreesCheck struct{}

func (worktreesCheck) Name() string { return "worktrees" }

func (worktreesCheck) Run(ctx Context) Diagnosis {
	d := Diagnosis{Name: "worktrees"}
	if !gitAvailable(ctx.Root) {
		d.Status = Warn
		d.Message = "no git context — cannot inspect worktrees"
		return d
	}
	abandoned := abandonedWorktrees(ctx.Root)
	if len(abandoned) == 0 {
		d.Status = OK
		d.Message = "no abandoned worktrees"
		return d
	}
	d.Status = Error
	d.Message = "abandoned worktree(s) — remove manually"
	var cmds []string
	for _, w := range abandoned {
		cmd := "git worktree remove " + filepath.Join(worktree.Dir, w)
		d.Details = append(d.Details, w+": "+cmd)
		cmds = append(cmds, cmd)
	}
	d.Repair = &Repair{Command: joinCommands(cmds)}
	return d
}

// abandonedWorktrees returns the sorted feature names of worktrees whose
// workflow is complete/absent or whose branch is merged into main.
func abandonedWorktrees(root string) []string {
	var out []string
	for _, feature := range listWorktrees(root) {
		if workflowDone(root, feature) || branchMerged(root, feature) {
			out = append(out, feature)
		}
	}
	sort.Strings(out)
	return out
}

// joinCommands joins remediation commands with " && " for a single copy-paste.
func joinCommands(cmds []string) string {
	out := ""
	for i, c := range cmds {
		if i > 0 {
			out += " && "
		}
		out += c
	}
	return out
}
