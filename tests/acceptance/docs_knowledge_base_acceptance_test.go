package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/docs-knowledge-base-pages.feature
func TestKnowledgeBaseScaffoldingIsWiredAndEnforced(t *testing.T) {
	root := filepath.Join("..", "..")

	loaderPath := filepath.Join(root, "internal", "docgen", "load_kb.go")
	loader, err := os.ReadFile(loaderPath)
	if err != nil {
		t.Fatalf("read loader: %v", err)
	}
	for _, want := range []string{"loadKBPages", "parseKBFile", "What it does", "When you'd use it", "How it behaves"} {
		if !strings.Contains(string(loader), want) {
			t.Fatalf("loader missing %q", want)
		}
	}

	renderPath := filepath.Join(root, "internal", "docgen", "render_kb.go")
	renderer, err := os.ReadFile(renderPath)
	if err != nil {
		t.Fatalf("read renderer: %v", err)
	}
	for _, want := range []string{"RenderKBIndex", "RenderKBFeature"} {
		if !strings.Contains(string(renderer), want) {
			t.Fatalf("renderer missing %q", want)
		}
	}

	navPath := filepath.Join(root, "internal", "docgen", "render_nav.go")
	nav, err := os.ReadFile(navPath)
	if err != nil {
		t.Fatalf("read nav: %v", err)
	}
	if !strings.Contains(string(nav), `kb/index.html`) {
		t.Fatal("main docs nav must link to kb/index.html")
	}

	validatorPath := filepath.Join(root, "internal", "workflow", "validate_docs.go")
	v, err := os.ReadFile(validatorPath)
	if err != nil {
		t.Fatalf("read validator: %v", err)
	}
	for _, want := range []string{"knowledge base markdown missing", "knowledge base page missing", "docs/project-docs/kb"} {
		if !strings.Contains(string(v), want) {
			t.Fatalf("validator missing %q", want)
		}
	}

	promptPath := filepath.Join(root, "docs", "architecture", "documentation-generator-prompt.md")
	p, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("read prompt: %v", err)
	}
	for _, want := range []string{"docs/project-docs/kb/<feature>.md", "Audience: Centinela end-users"} {
		if !strings.Contains(string(p), want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
}
