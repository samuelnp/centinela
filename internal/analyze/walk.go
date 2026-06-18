package analyze

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// skipDirs are directory names never descended into: dependency, VCS, build,
// and Centinela-internal trees. Counts then reflect real source (AC-5).
var skipDirs = map[string]bool{
	"vendor": true, "node_modules": true, ".git": true,
	".workflow": true, "dist": true, "build": true,
}

const maxLayoutDepth = 2

// walkResult is the read-only scan output: per-extension file counts and a
// depth-bounded, sorted list of package/directory paths for the layout.
type walkResult struct {
	extCounts map[string]int
	packages  []string
}

// walk traverses root read-only, skipping skipDirs and gitignored paths, never
// following directory symlinks, and tolerating unreadable entries (they are
// skipped, not fatal). It returns per-extension counts and a depth-bounded
// package layout. An unreadable root is the sole hard error.
func walk(root string) (walkResult, error) {
	if _, err := os.ReadDir(root); err != nil {
		return walkResult{}, err
	}
	ign := loadGitignore(root)
	res := walkResult{extCounts: map[string]int{}}
	pkgSet := map[string]bool{}
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil || rel == "." {
			return nil
		}
		if d.IsDir() {
			if skipDirs[d.Name()] || ign.match(rel) {
				return filepath.SkipDir
			}
			if depth(rel) <= maxLayoutDepth {
				pkgSet[filepath.ToSlash(rel)] = true
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 || ign.match(rel) {
			return nil
		}
		if ext := strings.ToLower(filepath.Ext(path)); ext != "" {
			res.extCounts[ext]++
		}
		return nil
	})
	for p := range pkgSet {
		res.packages = append(res.packages, p)
	}
	sort.Strings(res.packages)
	return res, nil
}

// depth counts path segments in a module-relative slash path.
func depth(rel string) int {
	return strings.Count(filepath.ToSlash(rel), "/") + 1
}
