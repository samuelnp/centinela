package main

import (
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/worktree"
)

// stewardEvidenceValidatorFor adapts orchestration.ValidateEvidence into the
// worktree.StewardEvidenceValidator signature, reading the evidence in the
// PRIMARY tree — where the stalled merge and its pending marker live — so the
// resume path works identically from the primary CWD and from a worktree CWD.
// cmd/ is the only layer allowed to import both internal/orchestration and
// internal/worktree, so the validator is injected here and the worktree layer
// stays G2-clean.
func stewardEvidenceValidatorFor(repo string) worktree.StewardEvidenceValidator {
	return func(feature string) (string, error) {
		var verdict string
		err := inDir(repo, func() error {
			path := orchestration.JSONPath(feature, orchestration.RoleMergeSteward)
			if err := orchestration.ValidateEvidence(path, feature, "merge", orchestration.RoleMergeSteward, nil); err != nil {
				return err
			}
			var e error
			verdict, e = readStewardHandoff(path)
			return e
		})
		return verdict, err
	}
}

// runMergeContinue resumes a stalled merge in repo (the primary tree). The
// success line is printed by the SAME reporter the direct merge path uses, so
// it can only appear after a verified ref advance and a registry-verified
// worktree removal — it used to be an unconditional Println.
func runMergeContinue(repo, feature string) error {
	res, err := worktree.ResolveMerge(repo, feature, stewardEvidenceValidatorFor(repo), mergeRemovalOpts()...)
	if err != nil {
		return reportMergeFailure(res.Outcome, "centinela merge --continue "+feature, err)
	}
	if res.Escalated {
		fmt.Fprintln(os.Stderr, ui.RenderMergeEscalated(feature))
		fmt.Fprintln(os.Stderr, res.EscalationNote)
		return fmt.Errorf("merge %q escalated by Merge Steward — worktree kept for review", feature)
	}
	return reportMergeSuccess(res.Outcome)
}
