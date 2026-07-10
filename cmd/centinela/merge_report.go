package main

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/worktree"
)

// reportMergeSuccess prints the delivery confirmation — but ONLY when the
// outcome carries proof: the ref advanced, or the branch was verifiably
// already merged. Merge hard-errors before returning an unverified outcome,
// so the final guard here is defense in depth: if it ever trips, refuse to
// claim success rather than print the fabricated line this feature killed.
func reportMergeSuccess(o worktree.MergeOutcome) error {
	if !o.RefAdvanced && !o.AlreadyMerged {
		return fmt.Errorf("merge of %q finished without verified ref advance — refusing to claim success", o.Feature)
	}
	if o.AlreadyMerged {
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Branch %q was already merged into %s — worktree cleaned up, nothing new delivered.", o.Feature, o.TargetBranch)))
		return nil
	}
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Merged %q into %s and removed its worktree.", o.Feature, o.TargetBranch)))
	return nil
}
