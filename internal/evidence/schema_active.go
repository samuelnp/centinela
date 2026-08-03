package evidence

import (
	"github.com/samuelnp/centinela/internal/workflow"
	"github.com/samuelnp/centinela/internal/worktreepath"
)

// ResolveActiveFeature returns the feature the CWD identifies UNAMBIGUOUSLY and
// the directory that feature's workflow state must be READ FROM, or "", "" when
// there is no single answer. Two signals, in order:
//
//  1. cwd is inside a `.worktrees/<feature>` checkout — a worktree names exactly
//     one feature, so no guessing is involved. root is that checkout's root.
//  2. otherwise, exactly ONE active workflow on disk. root is "" — that scan is
//     already relative to the CWD, so the CWD is where the state lives.
//
// Zero active workflows and two-or-more both resolve to "" — the caller then
// emits an obvious unfilled slot rather than a plausible-but-wrong value.
//
// Returning root is what makes signal 1 usable from a SUBDIRECTORY. Every
// derivation input — `workflow.Load`, `docs/features/<feature>.md` — is read
// through a CWD-relative path, so resolving the slug without also telling the
// caller where to stand derives from a FAILED state lookup and falls back to
// the static legacy chain: a wrong handoffTo printed beside a CORRECT slug,
// which reads as more authoritative than the placeholder it replaced.
//
// This deliberately does NOT reuse `cmd/centinela`'s activeWorkflow, which
// falls back to the most-recently-touched of several active workflows. That
// heuristic is fine for a cost-attribution label (a wrong label is cosmetic)
// and wrong here: the derived handoffTo is checked by the completion gate, so
// guessing among parallel sessions reproduces exactly the bug this resolves.
func ResolveActiveFeature(cwd string) (feature, root string) {
	if f, r := worktreepath.DetectFeature(cwd); f != "" {
		return f, r
	}
	if wfs := workflow.ActiveWorkflows(workflow.WorkflowDir); len(wfs) == 1 {
		return wfs[0].Feature, ""
	}
	return "", ""
}
