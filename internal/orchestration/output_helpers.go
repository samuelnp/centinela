package orchestration

import (
	"os"
	"path/filepath"
	"strings"
)

func existingOutputFiles(outputs []string) []string {
	files := []string{}
	for _, out := range outputs {
		path := normalizeOutputPath(out)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			files = append(files, path)
		}
	}
	return files
}

func missingOutputFiles(outputs []string) []string {
	missing := []string{}
	for _, out := range outputs {
		path := normalizeOutputPath(out)
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			missing = append(missing, out)
		}
	}
	return missing
}

func normalizeOutputPath(path string) string {
	clean := strings.TrimSpace(strings.TrimPrefix(path, "./"))
	return filepath.ToSlash(filepath.Clean(clean))
}

func hasPathPrefix(paths []string, prefix string) bool {
	for _, path := range paths {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func containsPath(paths []string, want string) bool {
	for _, path := range paths {
		if path == want {
			return true
		}
	}
	return false
}

func hasImplementationOutput(paths []string) bool {
	for _, path := range paths {
		if !hasPathPrefix([]string{path}, ".workflow/") && !hasPathPrefix([]string{path}, "tests/") &&
			!hasPathPrefix([]string{path}, "docs/features/") && !hasPathPrefix([]string{path}, "docs/plans/") &&
			!hasPathPrefix([]string{path}, "specs/") && !hasPathPrefix([]string{path}, "docs/project-docs/") {
			return true
		}
	}
	return false
}
