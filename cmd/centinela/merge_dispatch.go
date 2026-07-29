package main

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/worktree"
)

// dispatchSteward records the pending-merge marker and prints the
// CENTINELA DIRECTIVE so the orchestrator invokes the merge-steward
// subagent. The worktree is kept and the command exits non-zero so CI
// and the UserPromptSubmit hook surface the block.
//
// repo MUST be the primary working tree: the stalled merge lives there, so
// the marker must too. Writing it relative to the invoking CWD desynchronised
// the conflict from its marker and made `merge --continue` unresumable.
func dispatchSteward(repo string, o worktree.MergeOutcome) error {
	if err := worktree.WritePending(repo, o); err != nil {
		return err
	}
	fmt.Println(ui.RenderMergeStewardNeeded(o.Feature, o.StewardReason()))
	fmt.Println(o.StewardDirective())
	return fmt.Errorf("merge requires Merge Steward review — resume with: centinela merge --continue %s", o.Feature)
}
