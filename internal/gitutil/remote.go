// Package gitutil is a tiny leaf wrapper around `git` and the GitHub CLI that
// answers the questions `centinela deliver`/`complete` ask before offering a
// delivery path: does the repo have an `origin` remote, is `gh` available, and
// which delivery options does the {origin × worktree} matrix permit. It is the
// single shared seam so the directive and the command agree on what is valid.
// It depends only on the standard library + os/exec — nothing internal.
package gitutil

import (
	"os/exec"
	"strings"
)

// gitRun is overridable for tests; default runs `git` via os/exec in repo.
var gitRun = func(repo string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// HasOriginRemote reports whether repo has a remote named "origin". A
// non-zero exit from `git remote get-url origin` (the remote is simply
// absent) is the normal "no" answer and returns (false, nil); only a real
// exec failure (e.g. git missing) returns an error.
func HasOriginRemote(repo string) (bool, error) {
	out, err := gitRun(repo, "remote", "get-url", "origin")
	if err == nil {
		return strings.TrimSpace(out) != "", nil
	}
	if _, ok := err.(*exec.ExitError); ok {
		return false, nil
	}
	return false, err
}
