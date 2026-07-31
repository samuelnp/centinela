package orchestration

import (
	"fmt"
	"os"
	"strings"
)

func IsUserFacingFeature(feature string) bool {
	data, err := os.ReadFile(fmt.Sprintf("docs/features/%s.md", feature))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if normalizeSurface(line) == "user-facing" {
			return true
		}
	}
	return false
}

// normalizeSurface extracts the surface value from a brief line. Markdown
// briefs commonly wrap the declaration in a blockquote (`> surface: ...`) or a
// list item (`- surface: ...`, `* surface: ...`), so leading marker characters
// are stripped before matching; the line-start discipline is preserved — prose
// mentioning "surface:" mid-line never matches.
func normalizeSurface(line string) string {
	text := strings.ToLower(strings.TrimSpace(line))
	text = strings.TrimLeft(text, ">-* \t")
	if !strings.HasPrefix(text, "surface:") {
		return ""
	}
	value := strings.TrimSpace(strings.TrimPrefix(text, "surface:"))
	return strings.ReplaceAll(strings.ReplaceAll(value, "_", "-"), " ", "-")
}
