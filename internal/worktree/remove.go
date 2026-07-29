package worktree

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Remove tears down a feature worktree. force=true uses `--force`.
// Idempotent: if the worktree does not exist, returns nil.
//
// The target comes from git's registry when it knows one, so a worktree
// created or moved outside `.worktrees/<feature>` is really removed instead
// of being silently skipped (which then read as a successful removal).
func Remove(repo, feature string, force bool) error {
	rel := removeTarget(repo, feature)
	if rel == "" {
		return nil
	}
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

// removeTarget picks what to hand `git worktree remove`: the registered path
// when git knows one, else the conventional relative path when it exists on
// disk, else "" meaning there is nothing to remove.
func removeTarget(repo, feature string) string {
	if path, ok, err := registeredWorktree(repo, branchName(feature)); err == nil && ok {
		return path
	}
	if Exists(repo, feature) {
		return filepath.Join(Dir, feature)
	}
	return ""
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
