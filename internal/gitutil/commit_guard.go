package gitutil

import (
	"os"
	"path/filepath"
	"strings"
)

// inFlightMarkers maps a git state file to the operator-facing reason it
// blocks a partial commit. git refuses `commit -- <pathspec>` while any of
// these exist, and a mutation run during a conflict resolution is exactly when
// a surprise commit is least wanted.
var inFlightMarkers = []struct{ marker, reason string }{
	{"MERGE_HEAD", "merge in progress"},
	{"rebase-merge", "rebase in progress"},
	{"rebase-apply", "rebase in progress"},
	{"CHERRY_PICK_HEAD", "cherry-pick in progress"},
	{"REVERT_HEAD", "revert in progress"},
}

// IsRepo reports whether repo is inside a git working tree.
func IsRepo(repo string) bool {
	_, err := gitRun(repo, "rev-parse", "--git-dir")
	return err == nil
}

// HasHead reports whether repo has a resolvable HEAD commit. A repository with
// no commits cannot take a partial commit against HEAD.
func HasHead(repo string) bool {
	_, err := gitRun(repo, "rev-parse", "--verify", "HEAD")
	return err == nil
}

// DetachedHead reports whether repo's HEAD points straight at a commit rather
// than at a branch. `symbolic-ref` fails exactly in that case; it SUCCEEDS on an
// unborn branch, which is why HasHead is the separate check.
//
// A commit made on a detached HEAD is reachable from nothing: the next
// `git checkout` orphans it and the record it carried is gone. That is worse
// than not committing, so it is a blocking condition, not a success.
func DetachedHead(repo string) bool {
	_, err := gitRun(repo, "symbolic-ref", "-q", "HEAD")
	return err != nil
}

// InProgressOperation names the in-flight git operation blocking a partial
// commit in repo ("merge in progress", "rebase in progress", …), or "" when
// none is running.
func InProgressOperation(repo string) string {
	for _, m := range inFlightMarkers {
		if gitPathExists(repo, m.marker) {
			return m.reason
		}
	}
	return ""
}

// CommitBlockedReason returns the human-readable reason a pathspec commit
// cannot be made in repo right now, or "" when one can. It is the single
// preflight both CommitPaths and its callers consult, so the warning an
// operator sees always names the condition that actually stopped the commit.
func CommitBlockedReason(repo string) string {
	if !IsRepo(repo) {
		return "no git repository"
	}
	if !HasHead(repo) {
		return "no HEAD"
	}
	// An in-flight rebase or bisect ALSO detaches HEAD, so the more specific
	// operation name is reported first.
	if op := InProgressOperation(repo); op != "" {
		return op
	}
	if DetachedHead(repo) {
		return "detached HEAD"
	}
	return ""
}

// gitPathExists resolves name inside repo's git directory (via `rev-parse
// --git-path`, which is worktree-aware — a linked worktree keeps its own
// MERGE_HEAD) and reports whether it exists.
func gitPathExists(repo, name string) bool {
	out, err := gitRun(repo, "rev-parse", "--git-path", name)
	if err != nil {
		return false
	}
	p := strings.TrimSpace(out)
	if p == "" {
		return false
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(repo, p)
	}
	_, err = os.Stat(p)
	return err == nil
}
