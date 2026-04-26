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

func normalizeSurface(line string) string {
	text := strings.ToLower(strings.TrimSpace(line))
	if !strings.HasPrefix(text, "surface:") {
		return ""
	}
	value := strings.TrimSpace(strings.TrimPrefix(text, "surface:"))
	return strings.ReplaceAll(strings.ReplaceAll(value, "_", "-"), " ", "-")
}
