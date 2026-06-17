package gitdiff

// ChangedFilesStaged returns the set of files staged in the index
// (git diff --cached), the input to the pre-commit gate run. It takes no base
// ref (the index is compared against HEAD, or against the empty tree on the
// initial commit). On any git failure it returns
// (nil, Summary{Degrade: reason}, nil) so the caller degrades — never crashes,
// never false-blocks.
func (r *Resolver) ChangedFilesStaged() (*Set, Summary, error) {
	summary := Summary{Base: "STAGED"}

	out, err := r.Run("git", "diff", "--cached", "--name-only", "--diff-filter=ACMR")
	if err != nil {
		summary.Degrade = degradeReason(err, "STAGED")
		return nil, summary, nil
	}

	set := NewSet(splitNonEmpty(out))
	summary.Files = set.Len()
	return set, summary, nil
}
