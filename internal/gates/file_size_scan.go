package gates

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

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

// itoa converts an int to string without importing strconv.
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
