package docstring

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// skippedDirs are never descended into by Files.
var skippedDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true, "dist": true,
	"testdata": true, ".worktrees": true, "build": true, "target": true,
}

// Files walks the configured roots and returns every in-scope file, sorted.
// It is the whole-repo *report* scope behind `centinela docs lint --full`; the
// gate never calls it, because a gate that opens legacy files it did not ask
// anyone to change is a permanently-red validator.
func Files(opts Options) ([]string, error) {
	var out []string
	for _, root := range roots(opts) {
		if _, err := os.Stat(root); err != nil {
			continue
		}
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if skippedDirs[d.Name()] {
					return filepath.SkipDir
				}
				return nil
			}
			if InScope(p, opts) {
				out = append(out, filepath.ToSlash(p))
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(out)
	return out, nil
}
