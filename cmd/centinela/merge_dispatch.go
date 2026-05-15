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
func dispatchSteward(o worktree.MergeOutcome) error {
	if err := worktree.WritePending(".", o); err != nil {
		return err
	}
	fmt.Println(ui.RenderMergeStewardNeeded(o.Feature, o.StewardReason()))
	fmt.Println(o.StewardDirective())
	return fmt.Errorf("merge requires Merge Steward review — resume with: centinela merge --continue %s", o.Feature)
}
