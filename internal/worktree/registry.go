package worktree

import (
	"fmt"
	"strings"
)

// registeredWorktree returns the path git has registered for branch in repo,
// read from `git worktree list --porcelain` — the actual registry, not the
// `.worktrees/<feature>` path convention. A worktree created or moved outside
// the convention is invisible to a plain os.Stat, which made "removed its
// worktree" claimable while the worktree was still live.
func registeredWorktree(repo, branch string) (string, bool, error) {
	out, err := gitRunner(repo, "worktree", "list", "--porcelain")
	if err != nil {
		return "", false, fmt.Errorf("git worktree list failed: %s: %w",
			strings.TrimSpace(string(out)), err)
	}
	path, ok := findRegisteredWorktree(string(out), branch)
	return path, ok, nil
}

// findRegisteredWorktree scans porcelain blocks (blank-line separated) for the
// live worktree checked out on branch. A block carrying `prunable` is a stale
// administrative record, not a worktree on disk, so it never counts as present.
func findRegisteredWorktree(out, branch string) (string, bool) {
	want := "branch refs/heads/" + branch
	for _, block := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n\n") {
		var path string
		var match, prunable bool
		for _, line := range strings.Split(block, "\n") {
			switch {
			case strings.HasPrefix(line, "worktree "):
				path = strings.TrimPrefix(line, "worktree ")
			case line == want:
				match = true
			case line == "prunable" || strings.HasPrefix(line, "prunable "):
				prunable = true
			}
		}
		if match && !prunable && path != "" {
			return path, true
		}
	}
	return "", false
}
