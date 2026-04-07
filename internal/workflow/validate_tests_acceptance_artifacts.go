package workflow

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var blockComments = regexp.MustCompile(`(?s)/\*.*?\*/`)

func hasExecutableAcceptanceTests(suffix string) bool {
	found := false
	filepath.WalkDir("tests/acceptance", func(path string, d os.DirEntry, err error) error {
		if err != nil || found || d.IsDir() || !isRealTestArtifact(path) {
			return nil
		}
		if suffix != "" && !strings.HasSuffix(path, suffix) {
			return nil
		}
		if hasExecutableAcceptanceContent(path) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func hasExecutableAcceptanceContent(path string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	clean := strings.TrimSpace(blockComments.ReplaceAllString(string(b), ""))
	if clean == "" {
		return false
	}
	lines := []string{}
	for _, line := range strings.Split(clean, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "//") || strings.HasPrefix(t, "#") || strings.HasPrefix(t, "--") {
			continue
		}
		lines = append(lines, t)
	}
	body := strings.ToLower(strings.Join(lines, "\n"))
	if body == "" || isPlaceholderAcceptanceBody(body) {
		return false
	}
	return looksLikeExecutableTest(body)
}

func isPlaceholderAcceptanceBody(body string) bool {
	return strings.Contains(body, "if false") || strings.Contains(body, "t.skip(") || strings.Contains(body, "placeholder") || strings.Contains(body, "todo")
}

func looksLikeExecutableTest(body string) bool {
	keys := []string{"t.", "given(", "when(", "then(", "describe(", "it(", "scenario(", "def test", "assert", "expect(", "should "}
	for _, k := range keys {
		if strings.Contains(body, k) {
			return true
		}
	}
	return false
}
