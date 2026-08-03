package worktree

import (
	"os"
	"path/filepath"

	"github.com/samuelnp/centinela/internal/worktreepath"
)

// Dir is the directory under the repo root where feature worktrees live. The
// literal and the cwd scan live in the `worktreepath` leaf so packages outside
// this one's sanctioned consumers (PROJECT.md G2) can share them without
// importing this package.
const Dir = worktreepath.Dir

// Path returns the canonical worktree path for a feature inside repo.
func Path(repo, feature string) string {
	return filepath.Join(repo, Dir, feature)
}

// DetectFeatureFromCwd walks parents of cwd looking for a `.worktrees/<feature>`
// segment. Returns the feature slug and the worktree root, or empty strings
// when cwd is not inside any worktree. Resolves symlinks (e.g. macOS
// `/tmp` -> `/private/tmp`) before scanning so links along the path do not
// mask the `.worktrees/<feature>` segment. The scan itself lives in the
// `worktreepath` leaf; this is the lifecycle package's name for it.
func DetectFeatureFromCwd(cwd string) (feature, root string) {
	return worktreepath.DetectFeature(cwd)
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
