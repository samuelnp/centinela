package worktree

import (
	"fmt"
	"strings"
)

// headSHA returns the commit SHA of HEAD in the working tree at repo. It
// also fails fast on a repository with no commits, where merging is moot.
func headSHA(repo string) (string, error) {
	out, err := gitRunner(repo, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("cannot read HEAD in %s: %s: %w", repo, strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out)), nil
}

// currentBranch returns the branch checked out at repo. A detached HEAD
// (symbolic-ref failure) is an error: advancing a detached HEAD is not
// "main advanced", so the merge must refuse before touching anything.
func currentBranch(repo string) (string, error) {
	out, err := gitRunner(repo, "symbolic-ref", "-q", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("primary working tree %s has a detached HEAD — check out the target branch before merging", repo)
	}
	return strings.TrimSpace(string(out)), nil
}

// isAncestor reports whether branch is already an ancestor of HEAD at repo.
func isAncestor(repo, branch string) bool {
	_, err := gitRunner(repo, "merge-base", "--is-ancestor", branch, "HEAD")
	return err == nil
}

// verifyRemoved confirms the feature worktree directory is actually gone.
// Remove is idempotent-by-design, so a wrong-CWD no-op looks identical to a
// real removal unless the directory itself is checked.
func verifyRemoved(repo, feature string) error {
	if Exists(repo, feature) {
		return fmt.Errorf("worktree %s still exists after removal — not claiming success", Path(repo, feature))
	}
	return nil
}

// verifyAdvance asserts success on the ref, not on the absence of git
// errors. It re-reads HEAD and sets RefAdvanced when it moved past before;
// otherwise AlreadyMerged when the branch is already an ancestor of HEAD.
// Neither flag set means git exited 0 without delivering the branch — a
// hard error so the caller can never print a fabricated success line.
func verifyAdvance(o *MergeOutcome, repo, before string) error {
	after, err := headSHA(repo)
	if err != nil {
		return err
	}
	o.RefAdvanced = before != after
	if !o.RefAdvanced {
		o.AlreadyMerged = isAncestor(repo, o.Branch)
	}
	if !o.RefAdvanced && !o.AlreadyMerged {
		return fmt.Errorf("merge reported success but %q did not advance HEAD in %s", o.Branch, repo)
	}
	return nil
}
