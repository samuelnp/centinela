package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/docs-latest-features-getting-started.feature
// Scenario: Getting-started docs and the generated story stay in sync
// The getting-started prose moved from the README into docs/guides/getting-started.md
// when the README was slimmed to a landing page; the generated-docs story
// (render_story.go) still emits the Latest Features + Getting Started sections. This
// test keeps the source guide and the generated story in sync.
func TestDocsLatestFeaturesAndGettingStartedStayInSync(t *testing.T) {
	guide, err := os.ReadFile(filepath.Join("..", "..", "docs", "guides", "getting-started.md"))
	if err != nil {
		t.Fatalf("read getting-started guide: %v", err)
	}
	for _, want := range []string{"Getting Started", "centinela roadmap validate", "centinela migrate docs"} {
		if !strings.Contains(string(guide), want) {
			t.Fatalf("getting-started guide missing %q", want)
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
