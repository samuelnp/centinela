// Package treestate captures the exact tree state a verdict certified.
//
// A HEAD revision alone is not enough (D3): `centinela complete` auto-commits
// only AFTER its gates run, so an operator who fixes code in place after a
// blocking verdict leaves HEAD untouched. The snapshot therefore pairs HEAD
// with a digest of the working tree.
//
// The digest EXCLUDES everything under .workflow/ (D3a). The verifier writes
// its own report and evidence there, so including it would make every stamp
// stale the instant it was written.
package treestate

import (
	"fmt"
	"strings"
)

// Snapshot binds a verdict to a tree state.
type Snapshot struct {
	Revision string
	Digest   string
}

// Stamp reads HEAD and the working-tree delta through run and reduces them to
// a Snapshot. A repository without commits, or no git at all, is an error —
// freshness is never assumed.
func Stamp(root string, run Runner) (Snapshot, error) {
	revision, err := run(root, "rev-parse", "HEAD")
	if err != nil {
		return Snapshot{}, fmt.Errorf("reading HEAD revision: %w", err)
	}
	status, err := run(root, "status", "--porcelain=v1")
	if err != nil {
		return Snapshot{}, fmt.Errorf("reading working-tree status: %w", err)
	}
	diff, err := run(root, "diff", "HEAD")
	if err != nil {
		return Snapshot{}, fmt.Errorf("reading working-tree diff: %w", err)
	}
	return Snapshot{Revision: strings.TrimSpace(revision), Digest: Digest(status, diff)}, nil
}
