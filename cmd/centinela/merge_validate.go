package main

import (
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/config"
)

// inDir runs fn with the process CWD set to dir, restoring it afterwards — the
// one seam for work that must read the CWD-relative domain APIs from somewhere
// other than the invoking directory. Two callers:
//
//   - merge-time work (validate, portal regen, steward evidence lookup) must
//     operate on the PRIMARY tree: from a feature worktree the invoking CWD is
//     about to be deleted, and its diff against the target is empty because the
//     primary tree has already absorbed the branch;
//   - `evidence schema` must derive from the WORKTREE ROOT, so it answers the
//     same from a package subdirectory as from the root.
//
// A missing original CWD (the worktree we were standing in got removed) is
// tolerated — there is simply nothing to restore.
func inDir(dir string, fn func() error) error {
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("cannot enter %s: %w", dir, err)
	}
	if orig != "" {
		defer os.Chdir(orig) //nolint:errcheck // best-effort restore
	}
	return fn()
}

// runValidateForMerge validates the MERGED result in the primary tree.
// The scope is forced full: the diff base is the target branch, which by
// this point already contains the merge, so a diff-aware run would report
// "0 files changed" and pass vacuously — a gate that gates nothing.
func runValidateForMerge(repo string) (bool, string) {
	if err := inDir(repo, func() error {
		return executeValidationWithFlag(config.FlagForceFull)
	}); err != nil {
		return false, err.Error()
	}
	return true, ""
}
