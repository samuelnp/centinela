package worktree

import "fmt"

// StewardEvidenceValidator re-validates `.workflow/<feature>-merge-steward.json`
// and returns the steward verdict (the evidence `handoffTo` value:
// "complete" for APPLY, "user" for ESCALATE). A non-nil error means the
// evidence is missing or schema-invalid and finalization must be refused.
// Injected from cmd/ so this layer never imports internal/orchestration.
type StewardEvidenceValidator func(feature string) (verdict string, err error)

// Resolution is the outcome of a `centinela merge --continue` attempt.
type Resolution struct {
	Finalized      bool
	Escalated      bool
	Verdict        string
	EscalationNote string
	// Outcome carries the same verified flags the direct merge path produces
	// (RefAdvanced / AlreadyMerged / RemoveFailed / TargetBranch) so the CLI
	// routes --continue through the identical success reporter. Without it,
	// --continue printed the success line unconditionally.
	Outcome MergeOutcome
}

// ResolveMerge gates finalization of a stalled merge on steward evidence.
//
//   - no pending marker            -> error (nothing to continue)
//   - dirty main tree              -> error (blocked even on APPLY)
//   - invalid/missing evidence     -> error (validator message surfaced)
//   - valid + handoffTo "complete" -> verify the branch actually landed in
//     repo, then finalize: remove worktree, verify removal, clear marker
//   - valid + handoffTo "user"     -> escalate: keep worktree + marker
//
// repo MUST be the primary working tree (the tree the stalled merge lives
// in). An APPLY verdict is the steward's claim, not proof: finalization is
// gated on ancestry in repo so --continue can never print a success the ref
// does not back.
func ResolveMerge(repo, feature string, validate StewardEvidenceValidator, opts ...MergeOption) (Resolution, error) {
	var r Resolution
	m, err := LoadPending(repo, feature)
	if err != nil {
		return r, err
	}
	if m == nil {
		return r, fmt.Errorf("no pending merge to continue for %q", feature)
	}
	if dirty, err := isDirty(repo); err != nil {
		return r, err
	} else if dirty {
		return r, fmt.Errorf("main working tree is dirty — commit or stash before continuing %q", feature)
	}
	verdict, err := validate(feature)
	if err != nil {
		return r, fmt.Errorf("steward evidence required: %w", err)
	}
	r.Verdict = verdict
	if verdict != "complete" {
		r.Escalated = true
		r.EscalationNote = stewardEscalationDetail(repo, feature)
		return r, nil
	}
	o := MergeOutcome{Feature: feature, Branch: branchName(feature)}
	if o.TargetBranch, err = currentBranch(repo); err != nil {
		return r, err
	}
	if err := verifyLanded(&o, repo, m.BaseSHA); err != nil {
		return r, err
	}
	if err := finishMerge(&o, repo, resolveMergeOpts(opts)); err != nil {
		r.Outcome = o
		return r, err
	}
	if err := ClearPending(repo, feature); err != nil {
		return r, err
	}
	r.Outcome, r.Finalized = o, true
	return r, nil
}
