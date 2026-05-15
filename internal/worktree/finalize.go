package worktree

import (
	"fmt"
	"os"
)

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
}

// ResolveMerge gates finalization of a stalled merge on steward evidence.
//
//   - no pending marker            -> error (nothing to continue)
//   - dirty main tree              -> error (blocked even on APPLY)
//   - invalid/missing evidence     -> error (validator message surfaced)
//   - valid + handoffTo "complete" -> finalize: remove worktree, clear marker
//   - valid + handoffTo "user"     -> escalate: keep worktree + marker
func ResolveMerge(repo, feature string, validate StewardEvidenceValidator) (Resolution, error) {
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
	if err := Remove(repo, feature, false); err != nil {
		return r, err
	}
	if err := ClearPending(repo, feature); err != nil {
		return r, err
	}
	r.Finalized = true
	return r, nil
}

// stewardEscalationDetail returns the steward report plus its proposed
// diff sibling (when present) so the caller can surface them to stderr.
func stewardEscalationDetail(repo, feature string) string {
	o := MergeOutcome{Feature: feature}
	detail := readFileOr(o.StewardHint(), "(no steward report found)")
	if diff, err := os.ReadFile(fmt.Sprintf(".workflow/%s-merge-steward.diff", feature)); err == nil {
		detail += "\n\n--- proposed diff ---\n" + string(diff)
	}
	return detail
}

func readFileOr(path, fallback string) string {
	if data, err := os.ReadFile(path); err == nil {
		return string(data)
	}
	return fallback
}
