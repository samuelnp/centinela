package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadmeCentinelaUsageHighlightsCurrentFeatures(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	readme := string(data)

	for _, want := range []string{
		"Plan advisor mode",
		"Actionable specialist orchestration",
		"Stronger quality gates",
		"HOWTO.md",
	} {
		if !strings.Contains(readme, want) {
			t.Fatalf("README missing %q", want)
		}
	}
}
