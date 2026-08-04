// Package worktreepath is the `.worktrees/<feature>` path vocabulary: the
// directory name and the cwd -> feature scan, nothing else.
//
// It is a LEAF (standard library only) on purpose. `internal/worktree` (the
// worktree lifecycle domain) and `internal/evidence` (skeleton feature
// resolution) both need the scan; exporting it from `internal/worktree` would
// force an `internal/evidence -> internal/worktree` edge that PROJECT.md G2
// does not sanction, while a stdlib-only leaf can be shared by both and cannot
// participate in a cycle. Same reasoning as `internal/acceptance` and
// `internal/docstring`.
package worktreepath

import (
	"path/filepath"
	"strings"
)

// Dir is the directory under the repo root where feature worktrees live.
const Dir = ".worktrees"

// DetectFeature walks parents of cwd looking for a `.worktrees/<feature>`
// segment. Returns the feature slug and the worktree root, or empty strings
// when cwd is not inside any worktree. Resolves symlinks (e.g. macOS
// `/tmp` -> `/private/tmp`) before scanning so links along the path do not
// mask the `.worktrees/<feature>` segment.
func DetectFeature(cwd string) (feature, root string) {
	// An Abs failure (the process cwd was deleted) yields "", whose scan finds
	// no segment and returns no feature — the same answer an explicit error
	// branch would give, without an arm no test can reach.
	abs, _ := filepath.Abs(cwd)
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
