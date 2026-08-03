package gitdiff

import (
	"fmt"
	"strings"
)

// remoteRef returns the remote-tracking form of a bare branch name, or "" when
// base is already qualified (contains a slash) and needs no rewriting.
func remoteRef(base string) string {
	if base == "" || strings.Contains(base, "/") {
		return ""
	}
	return defaultRemote + "/" + base
}

// defaultRemote is the remote whose tracking refs are tried when a bare branch
// name does not resolve locally. "origin" is git's own clone default.
const defaultRemote = "origin"

// mergeBase resolves base against HEAD, retrying the remote-tracking form when
// the bare name does not exist locally. That retry is load-bearing, not a
// convenience: `actions/checkout` leaves a detached HEAD with no local branch
// ref for the default branch, so `git merge-base HEAD main` exits 128 with
// "Not a valid object name" even on a full-history clone. Without the retry
// every diff-aware gate silently degrades to "base not found" in CI — a gate
// that never runs looks exactly like a gate that always passes.
//
// It returns the ref that actually resolved (so callers can report the truth),
// the merge-base commit, and the original error when neither form resolves.
func (r *Resolver) mergeBase(base string) (string, string, error) {
	out, err := r.Run("git", "merge-base", "HEAD", base)
	if err == nil {
		return base, strings.TrimSpace(out), nil
	}
	remote := remoteRef(base)
	if remote == "" {
		return base, "", err
	}
	remoteOut, remoteErr := r.Run("git", "merge-base", "HEAD", remote)
	if remoteErr == nil {
		return remote, strings.TrimSpace(remoteOut), nil
	}
	return base, "", err
}

// degradeReason maps a failed base resolution to a human reason, naming every
// ref form that was tried so the notice cannot imply only one was attempted.
func degradeReason(err error, base string) string {
	msg := err.Error()
	if strings.Contains(msg, "not a git repository") {
		return "not a git repository"
	}
	if strings.Contains(msg, "Not a valid object name") ||
		strings.Contains(msg, "unknown revision") {
		if remote := remoteRef(base); remote != "" {
			return fmt.Sprintf("diff base %q not found (also tried %q)", base, remote)
		}
		return fmt.Sprintf("diff base %q not found", base)
	}
	return fmt.Sprintf("git merge-base failed: %s", msg)
}
