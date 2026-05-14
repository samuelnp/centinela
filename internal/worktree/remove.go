package worktree

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Remove tears down a feature worktree. force=true uses `--force`.
// Idempotent: if the worktree does not exist, returns nil.
func Remove(repo, feature string, force bool) error {
	if !Exists(repo, feature) {
		return nil
	}
	rel := filepath.Join(Dir, feature)
	args := []string{"worktree", "remove", rel}
	if force {
		args = append(args, "--force")
	}
	out, err := gitRunner(repo, args...)
	if err != nil {
		return fmt.Errorf("git worktree remove failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// DeleteBranch removes the local feature branch after the worktree is gone.
// Errors are non-fatal callers that still want to surface them can inspect.
func DeleteBranch(repo, feature string) error {
	out, err := gitRunner(repo, "branch", "-D", branchName(feature))
	if err != nil {
		return fmt.Errorf("git branch -D failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}
