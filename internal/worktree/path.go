package worktree

import (
	"os"
	"path/filepath"
	"strings"
)

// Dir is the directory under the repo root where feature worktrees live.
const Dir = ".worktrees"

// Path returns the canonical worktree path for a feature inside repo.
func Path(repo, feature string) string {
	return filepath.Join(repo, Dir, feature)
}

// DetectFeatureFromCwd walks parents of cwd looking for a `.worktrees/<feature>`
// segment. Returns the feature slug and the worktree root, or empty strings
// when cwd is not inside any worktree. Resolves symlinks (e.g. macOS
// `/tmp` -> `/private/tmp`) before scanning so links along the path do not
// mask the `.worktrees/<feature>` segment.
func DetectFeatureFromCwd(cwd string) (feature, root string) {
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return "", ""
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	parts := strings.Split(filepath.ToSlash(abs), "/")
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == Dir {
			feature = parts[i+1]
			root = filepath.FromSlash(strings.Join(parts[:i+2], "/"))
			return feature, root
		}
	}
	return "", ""
}

// IsInsideWorktree reports whether cwd is inside any `.worktrees/<feature>`.
func IsInsideWorktree(cwd string) bool {
	feature, _ := DetectFeatureFromCwd(cwd)
	return feature != ""
}

// Exists reports whether the worktree directory for feature exists on disk.
func Exists(repo, feature string) bool {
	info, err := os.Stat(Path(repo, feature))
	return err == nil && info.IsDir()
}
