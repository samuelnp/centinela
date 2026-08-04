package gitutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrNothingToCommit reports that none of the requested paths differ from
// HEAD, so no commit was made. It is a normal outcome, not a failure: a
// mutation that leaves state byte-identical must create no empty commit.
var ErrNothingToCommit = errors.New("nothing to commit")

// CommitPaths creates ONE commit in repo containing exactly paths and nothing
// else, with `--no-verify`.
//
// It is a *partial* commit: git composes the new tree from HEAD plus the
// pathspec, so the index and working tree for every other path are left
// byte-identical — unrelated staged work stays staged and uncommitted. That is
// the whole reason this exists instead of `git add -A` + commit.
//
// --no-verify is deliberate and scoped to generated governance state: the one
// gate that governs these files (roadmap_drift) is clean by construction
// because the caller just regenerated them, and running the full precommit
// suite on every `roadmap defer` would make deferring cost a full validate.
//
// Paths that exist neither on disk nor in the index are dropped rather than
// failing the commit, so a declared-but-unwritten artifact path is harmless.
func CommitPaths(repo, msg string, paths []string) error {
	if reason := CommitBlockedReason(repo); reason != "" {
		return errors.New(reason)
	}
	present := existingPaths(repo, paths)
	if len(present) == 0 {
		return ErrNothingToCommit
	}
	// Stage first so a brand-new (untracked) roadmap file is known to git;
	// `commit -- <pathspec>` only matches paths git already tracks.
	if out, err := gitRun(repo, append([]string{"add", "--"}, present...)...); err != nil {
		return fmt.Errorf("git add failed: %s", firstLine(out, err))
	}
	changed, err := gitRun(repo, append([]string{"status", "--porcelain=v1", "--"}, present...)...)
	if err != nil {
		return fmt.Errorf("git status failed: %s", firstLine(changed, err))
	}
	if strings.TrimSpace(changed) == "" {
		return ErrNothingToCommit
	}
	args := append([]string{"commit", "--no-verify", "-q", "-m", msg, "--"}, present...)
	if out, err := gitRun(repo, args...); err != nil {
		return fmt.Errorf("commit failed: %s", firstLine(out, err))
	}
	return nil
}

// existingPaths reduces a declared pathspec to the entries git can actually
// commit. A path deleted by the mutation is KEPT when git still tracks it, so a
// removal is committed rather than dropped; a path the project gitignores and
// does not track is DROPPED, because `git add` would refuse the whole pathspec
// over it and take the real mutation's commit down with it.
func existingPaths(repo string, paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		tracked := isTracked(repo, p)
		if !tracked && isIgnored(repo, p) {
			continue
		}
		if _, err := os.Stat(filepath.Join(repo, p)); err == nil || tracked {
			out = append(out, p)
		}
	}
	return out
}

// isTracked reports whether git has p in its index.
func isTracked(repo, p string) bool {
	out, err := gitRun(repo, "ls-files", "--error-unmatch", "--", p)
	return err == nil && out != ""
}

// isIgnored reports whether p matches an ignore rule. `check-ignore` exits 0
// only on a match, so a non-zero exit is the ordinary "not ignored" answer.
func isIgnored(repo, p string) bool {
	_, err := gitRun(repo, "check-ignore", "-q", "--", p)
	return err == nil
}

// firstLine reduces git's combined output to one reportable line, falling back
// to the exec error when git said nothing.
func firstLine(out string, err error) string {
	if s := strings.TrimSpace(strings.SplitN(strings.TrimSpace(out), "\n", 2)[0]); s != "" {
		return s
	}
	return err.Error()
}
