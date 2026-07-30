package gates

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

const maxLines = 100

var sourceRoots = []string{"src", "internal", "cmd", "lib", "app", "pkg"}
var ignoreDirs = []string{".git", "node_modules", "vendor", "dist", ".next", "target", "build"}

// checkFileSize reports what it actually inspected. An empty diff scope means
// zero files were read, which is a Skip — a green ✓ for a gate that inspected
// nothing is a false assurance. The cap and the Fail severity are unchanged.
func checkFileSize(cfg *config.Config, filter *gitdiff.Set) Result {
	if filter != nil && filter.Len() == 0 {
		return Result{Name: fileSizeGate, Status: Skip, Message: fileSizeNothingInspected}
	}
	violations, justified := findOversizedFiles(cfg, filter)
	if len(violations) == 0 {
		return Result{
			Name:    fileSizeGate,
			Status:  Pass,
			Message: fileSizePassMessage(justified),
			Details: justified,
		}
	}
	return Result{
		Name:    fileSizeGate,
		Status:  Fail,
		Message: fileSizeFailMessage(),
		Details: violations,
	}
}

func findOversizedFiles(cfg *config.Config, filter *gitdiff.Set) ([]string, []string) {
	roots := existingRoots()
	if len(roots) == 0 {
		roots = []string{"."}
	}
	exceptions := fileSizeExceptionMap(cfg)

	seen := map[string]bool{}
	var violations []string
	var justified []string

	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if shouldSkipDir(d.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			if seen[path] || !isSourceFile(path) {
				return nil
			}
			seen[path] = true
			if filter != nil && !filter.Contains(path) {
				return nil
			}

			if n := countLines(path); n > maxLines {
				rel := filepath.ToSlash(path)
				if ex, ok := exceptions[rel]; ok {
					if n <= ex.MaxLines {
						justified = append(justified, fmt.Sprintf("%s (%d lines) justified as %s: %s", rel, n, ex.Kind, ex.Reason))
						return nil
					}
					violations = append(violations, fmt.Sprintf("%s (%d lines) exceeds justified max %d", rel, n, ex.MaxLines))
					return nil
				}
				violations = append(violations, formatViolation(path, n))
			}
			return nil
		})
	}
	return violations, justified
}

func existingRoots() []string {
	var out []string
	for _, r := range sourceRoots {
		if _, err := os.Stat(r); err == nil {
			out = append(out, r)
		}
	}
	return out
}
