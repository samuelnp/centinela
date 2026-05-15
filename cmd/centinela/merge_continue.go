package main

import (
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/worktree"
)

// stewardEvidenceValidator adapts orchestration.ValidateEvidence into the
// worktree.StewardEvidenceValidator signature. cmd/ is the only layer
// allowed to import both internal/orchestration and internal/worktree, so
// the validator is injected here and the worktree layer stays G2-clean.
func stewardEvidenceValidator(feature string) (string, error) {
	path := orchestration.JSONPath(feature, orchestration.RoleMergeSteward)
	if err := orchestration.ValidateEvidence(path, feature, "merge", orchestration.RoleMergeSteward, nil); err != nil {
		return "", err
	}
	return readStewardHandoff(path)
}

func runMergeContinue(feature string) error {
	res, err := worktree.ResolveMerge(".", feature, stewardEvidenceValidator)
	if err != nil {
		return err
	}
	if res.Escalated {
		fmt.Fprintln(os.Stderr, ui.RenderMergeEscalated(feature))
		fmt.Fprintln(os.Stderr, res.EscalationNote)
		return fmt.Errorf("merge %q escalated by Merge Steward — worktree kept for review", feature)
	}
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Merged %q into main and removed its worktree.", feature)))
	return nil
}
