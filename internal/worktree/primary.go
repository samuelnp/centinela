package worktree

import (
	"fmt"
	"strings"
)

// PrimaryTree resolves the primary working tree of the repository containing
// cwd by parsing `git worktree list --porcelain`. By git contract the first
// `worktree <path>` block is the primary tree, so a merge run from inside a
// feature worktree can still target the right checkout. It never guesses:
// runner failures, empty or unparseable output, and a bare primary tree are
// all hard errors.
func PrimaryTree(cwd string) (string, error) {
	out, err := gitRunner(cwd, "worktree", "list", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("cannot resolve primary working tree: git worktree list failed: %s: %w",
			strings.TrimSpace(string(out)), err)
	}
	path, bare, found := parseFirstWorktreeBlock(string(out))
	if !found {
		return "", fmt.Errorf("cannot resolve primary working tree: no worktree entry in `git worktree list --porcelain` output")
	}
	if bare {
		return "", fmt.Errorf("cannot resolve primary working tree: primary working tree is bare")
	}
	return path, nil
}

// parseFirstWorktreeBlock scans porcelain output for the first worktree
// block. The path is everything after the first space (paths may contain
// spaces); a blank line terminates the block, whose remaining attribute
// lines are checked for the `bare` marker. Trailing blank lines and \r
// line endings are tolerated.
func parseFirstWorktreeBlock(out string) (path string, bare, found bool) {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			if found {
				return path, bare, true
			}
			continue
		}
		if !found {
			if strings.HasPrefix(line, "worktree ") {
				path = strings.TrimPrefix(line, "worktree ")
				found = true
			}
			continue
		}
		if line == "bare" {
			bare = true
		}
	}
	return path, bare, found
}
