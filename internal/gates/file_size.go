package gates

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const maxLines = 100

var sourceRoots = []string{"src", "internal", "cmd", "lib", "app", "pkg"}
var ignoreDirs = []string{".git", "node_modules", "vendor", "dist", ".next", "target", "build"}

func checkFileSize() Result {
	violations := findOversizedFiles()
	if len(violations) == 0 {
		return Result{Name: "G1: File Size", Status: Pass, Message: "All files under 100 lines."}
	}
	return Result{
		Name:    "G1: File Size",
		Status:  Fail,
		Message: "Files exceeding 100 lines must be split.",
		Details: violations,
	}
}

func findOversizedFiles() []string {
	roots := existingRoots()
	if len(roots) == 0 {
		roots = []string{"."}
	}

	seen := map[string]bool{}
	var violations []string

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

			if n := countLines(path); n > maxLines {
				violations = append(violations, formatViolation(path, n))
			}
			return nil
		})
	}
	return violations
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

func shouldSkipDir(name string) bool {
	for _, d := range ignoreDirs {
		if name == d {
			return true
		}
	}
	return strings.HasPrefix(name, ".")
}

func isSourceFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".rb", ".rs", ".java",
		".kt", ".cs", ".cpp", ".c", ".h", ".swift", ".gd":
		return true
	}
	return false
}

func countLines(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	n := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		n++
	}
	return n
}

func formatViolation(path string, lines int) string {
	return filepath.ToSlash(path) + " (" + itoa(lines) + " lines)"
}

// itoa converts an int to string without importing strconv (avoids extra import).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
