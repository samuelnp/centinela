package worktree

import (
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/config"
)

// MaybeProvision creates a worktree and chdirs into it when cfg requests it.
// It returns the worktree path it switched into, or "" when the flag is off
// or no git repo is present at repo. Errors are returned only for explicit
// failures (e.g. a git worktree call that did not succeed).
func MaybeProvision(repo, feature string, cfg *config.Config) (string, error) {
	if cfg == nil || !cfg.Workflow.UseWorktrees {
		return "", nil
	}
	if !isGitRepo(repo) {
		return "", nil
	}
	path, err := Create(repo, feature)
	if err != nil {
		return "", err
	}
	if err := os.Chdir(path); err != nil {
		return "", fmt.Errorf("worktree: cannot chdir into %s: %w", path, err)
	}
	return path, nil
}

func isGitRepo(repo string) bool {
	_, err := gitRunner(repo, "rev-parse", "--git-dir")
	return err == nil
}
