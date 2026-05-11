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

func checkFileSize(cfg *config.Config, filter *gitdiff.Set) Result {
	violations, justified := findOversizedFiles(cfg, filter)
	if len(violations) == 0 {
		msg := "All files under 100 lines."
		if filter != nil && filter.Len() == 0 {
			msg = "No relevant changes — gate skipped."
		} else if len(justified) > 0 {
			msg = "All files meet G1 (including justified exceptions)."
		}
		return Result{Name: "G1: File Size", Status: Pass, Message: msg, Details: justified}
	}
	return Result{
		Name:    "G1: File Size",
		Status:  Fail,
		Message: "Files exceeding 100 lines must be split unless explicitly justified.",
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
