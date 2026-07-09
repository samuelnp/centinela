package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReadmeCentinelaUsageHighlightsCurrentFeatures verifies the README landing
// page still reflects the current product: it states the enforced workflow, shows
// how Centinela works, links to the docs/guides, and points at HOWTO.md. The
// per-feature highlight prose that used to live inline in the README moved into
// docs/guides/ when the README was slimmed to a landing page, so this test asserts
// the landing page stays current rather than requiring the old inline copy.
func TestReadmeCentinelaUsageHighlightsCurrentFeatures(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	readme := string(data)

	for _, want := range []string{
		"plan → code → tests → validate → docs",
		"How Centinela Works",
		"docs/guides/getting-started.md",
		"HOWTO.md",
	} {
		if !strings.Contains(readme, want) {
			t.Fatalf("README missing %q", want)
		}
	}
}
