package worktree

import (
	"os"
	"path/filepath"
)

// ignoreEntry is the line appended to every ignore-style file.
const ignoreEntry = ".worktrees/"

// ignoreFilenames lists every plain-text ignore file we patch.
var ignoreFilenames = []string{
	".gitignore",
	".eslintignore",
	".prettierignore",
	".dockerignore",
	".rgignore",
}

// SyncResult records files that were patched. Order is deterministic.
type SyncResult struct {
	Touched []string
}

// SyncIgnores adds `.worktrees/` to every supported ignore file under repo,
// and patches `tsconfig.json` "exclude" if it exists. Idempotent: re-running
// adds nothing new and returns an empty Touched slice on the second call.
func SyncIgnores(repo string) (SyncResult, error) {
	var result SyncResult
	for _, name := range ignoreFilenames {
		changed, err := appendIgnoreLine(filepath.Join(repo, name), ignoreEntry)
		if err != nil {
			return result, err
		}
		if changed {
			result.Touched = append(result.Touched, name)
		}
	}
	changed, err := patchTsconfigExclude(filepath.Join(repo, "tsconfig.json"), ".worktrees")
	if err != nil {
		return result, err
	}
	if changed {
		result.Touched = append(result.Touched, "tsconfig.json")
	}
	return result, nil
}

// ensureFile creates an empty file when missing. Returns true if it was created.
func ensureFile(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		return false, err
	}
	return true, nil
}
