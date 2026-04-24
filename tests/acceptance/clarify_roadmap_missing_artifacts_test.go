package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/clarify-roadmap-missing-artifacts.feature
func TestArtifactTemplatesAndRoadmapRecoveryGuidanceExist(t *testing.T) {
	guide, err := os.ReadFile(filepath.Join("..", "..", "internal", "scaffold", "assets", "docs", "architecture", "new-project-guide.md"))
	if err != nil {
		t.Fatalf("read new project guide: %v", err)
	}
	for _, want := range []string{"centinela roadmap validate", "artifact-templates.md", ".workflow/roadmap.json"} {
		if !strings.Contains(string(guide), want) {
			t.Fatalf("new project guide missing %q", want)
		}
	}
	templates, err := os.ReadFile(filepath.Join("..", "..", "internal", "scaffold", "assets", "docs", "architecture", "artifact-templates.md"))
	if err != nil {
		t.Fatalf("read artifact templates: %v", err)
	}
	for _, want := range []string{".workflow/roadmap.json", ".workflow/<feature>-gatekeeper.md", ".workflow/<feature>-<role>.json"} {
		if !strings.Contains(string(templates), want) {
			t.Fatalf("artifact templates missing %q", want)
		}
	}
}
