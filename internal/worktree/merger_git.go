package worktree

import (
	"bufio"
	"bytes"
	"strings"
)

// isDirty reports whether the working tree at repo has uncommitted changes.
func isDirty(repo string) (bool, error) {
	out, err := gitRunner(repo, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return len(bytes.TrimSpace(out)) > 0, nil
}

// parseConflictedPaths returns paths git marks as unmerged.
func parseConflictedPaths(repo string) []string {
	out, err := gitRunner(repo, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil
	}
	var paths []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		p := strings.TrimSpace(scanner.Text())
		if p != "" {
			paths = append(paths, p)
		}
	}
	return paths
}
