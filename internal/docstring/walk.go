package docstring

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Files walks the configured roots and returns every in-scope file, sorted.
// Pruning here is an optimization only — InScope applies the same exclusion
// set, so the report and the gate see an identical file universe.
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
				if excludedDirs[d.Name()] {
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
