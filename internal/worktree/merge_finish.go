package worktree

// MergeOption tunes the merge sequence. Variadic so existing callers are
// unaffected.
type MergeOption func(*mergeOpts)

type mergeOpts struct{ forceRemove bool }

// WithForceRemove passes --force to `git worktree remove`, the escape hatch
// for a worktree git refuses to remove because it holds modified or untracked
// files. Only reachable behind an explicit operator flag: it discards those
// files, so it must never be the default.
func WithForceRemove() MergeOption {
	return func(o *mergeOpts) { o.forceRemove = true }
}

func resolveMergeOpts(opts []MergeOption) mergeOpts {
	var mo mergeOpts
	for _, f := range opts {
		f(&mo)
	}
	return mo
}

// finishMerge removes the delivered worktree and verifies it is really gone.
// A failure here is a HALF-success: the ref has already advanced, so it flags
// RemoveFailed on the outcome. Callers must report both halves — "merged, but
// removal failed" — instead of a bare exit 1 that hides the landed merge.
func finishMerge(o *MergeOutcome, repo string, mo mergeOpts) error {
	if err := Remove(repo, o.Feature, mo.forceRemove); err != nil {
		o.RemoveFailed = true
		return err
	}
	if err := verifyRemoved(repo, o.Feature); err != nil {
		o.RemoveFailed = true
		return err
	}
	return nil
}
