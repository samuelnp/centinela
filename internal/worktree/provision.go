package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// gitRunner is overridable for tests; default uses os/exec.
var gitRunner = func(repo string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	return cmd.CombinedOutput()
}

// Create provisions `.worktrees/<feature>` for the feature branch.
// Idempotent: if the worktree path already exists, returns it without error.
func Create(repo, feature string) (string, error) {
	if err := ValidateFeatureSlug(feature); err != nil {
		return "", err
	}
	target := Path(repo, feature)
	if Exists(repo, feature) {
		return target, nil
	}
	if err := os.MkdirAll(filepath.Join(repo, Dir), 0755); err != nil {
		return "", fmt.Errorf("worktree: cannot create parent dir: %w", err)
	}
	rel := filepath.Join(Dir, feature)
	args := []string{"worktree", "add", rel, "-b", branchName(feature)}
	if branchExists(repo, feature) {
		args = []string{"worktree", "add", rel, branchName(feature)}
	}
	out, err := gitRunner(repo, args...)
	if err != nil {
		return "", fmt.Errorf("git worktree add failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return target, nil
}

// branchName derives the git branch from a feature slug. v1: identity.
func branchName(feature string) string {
	return feature
}

func branchExists(repo, feature string) bool {
	_, err := gitRunner(repo, "rev-parse", "--verify", "refs/heads/"+branchName(feature))
	return err == nil
}
