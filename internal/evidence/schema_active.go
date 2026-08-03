package evidence

import (
	"github.com/samuelnp/centinela/internal/workflow"
	"github.com/samuelnp/centinela/internal/worktreepath"
)

// ResolveActiveFeature returns the feature the CWD identifies UNAMBIGUOUSLY, or
// "" when there is no single answer. Two signals, in order:
//
//  1. cwd is inside a `.worktrees/<feature>` checkout — a worktree names exactly
//     one feature, so no guessing is involved.
//  2. otherwise, exactly ONE active workflow on disk.
//
// Zero active workflows and two-or-more both resolve to "" — the caller then
// emits an obvious unfilled slot rather than a plausible-but-wrong value.
//
// This deliberately does NOT reuse `cmd/centinela`'s activeWorkflow, which
// falls back to the most-recently-touched of several active workflows. That
// heuristic is fine for a cost-attribution label (a wrong label is cosmetic)
// and wrong here: the derived handoffTo is checked by the completion gate, so
// guessing among parallel sessions reproduces exactly the bug this resolves.
func ResolveActiveFeature(cwd string) string {
	if feature, _ := worktreepath.DetectFeature(cwd); feature != "" {
		return feature
	}
	if wfs := workflow.ActiveWorkflows(workflow.WorkflowDir); len(wfs) == 1 {
		return wfs[0].Feature
	}
	return ""
}
