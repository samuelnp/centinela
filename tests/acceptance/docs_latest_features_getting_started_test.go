package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/docs-latest-features-getting-started.feature
func TestDocsLatestFeaturesAndGettingStartedStayInSync(t *testing.T) {
	readme, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	for _, want := range []string{"## Latest Features", "## Getting Started", "centinela roadmap validate", "centinela migrate docs", "2/5"} {
		if !strings.Contains(string(readme), want) {
			t.Fatalf("README missing %q", want)
		}
	}
	story, err := os.ReadFile(filepath.Join("..", "..", "internal", "docgen", "render_story.go"))
	if err != nil {
		t.Fatalf("read render story: %v", err)
	}
	for _, want := range []string{"Latest Features", "Getting Started", "centinela roadmap validate", "centinela docs generate --out docs/project-docs/index.html"} {
		if !strings.Contains(string(story), want) {
			t.Fatalf("render story missing %q", want)
		}
	}
}
