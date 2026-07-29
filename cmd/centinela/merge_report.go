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

// reportMergeFailure states what IS true when the merge itself landed but a
// later step failed. Exiting 1 with only the removal error told the operator
// nothing about the advanced ref — the mirror image of the false success this
// feature exists to kill — and left every re-run repeating the same failure.
// retry is the command that resumes from here.
func reportMergeFailure(o worktree.MergeOutcome, retry string, err error) error {
	if !o.RemoveFailed || (!o.RefAdvanced && !o.AlreadyMerged) {
		return err
	}
	landed := "merged into " + o.TargetBranch + " — verified"
	if o.AlreadyMerged {
		landed = "already merged into " + o.TargetBranch + " — verified"
	}
	return fmt.Errorf("%q %s; worktree removal failed: %w; re-run `%s --force-remove` to retry removal",
		o.Feature, landed, err, retry)
}
