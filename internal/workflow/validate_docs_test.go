package workflow

import (
	"os"
	"strings"
	"testing"
)

// The docs artifact gate is markdown-first: EVERY feature (user-facing
// included) requires a non-empty changelog; the real-updated-doc-file rule for
// user-facing features lives in orchestration evidence validation, not here.
func TestValidateDocsOutput(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                          //nolint:errcheck
	os.Chdir(d)                                                                //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                         //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("surface: user-facing\n"), 0644) //nolint:errcheck
	os.MkdirAll(WorkflowDir, 0755)                                             //nolint:errcheck
	if err := validateDocsOutput(""); err == nil {
		t.Fatal("expected missing feature error")
	}
	// A user-facing feature with no changelog fails naming the changelog path —
	// no knowledge-base bundle or portal is ever demanded any more.
	err := validateDocsOutput("f")
	if err == nil || !strings.Contains(err.Error(), "changelog entry missing") {
		t.Fatalf("expected changelog missing error, got %v", err)
	}
	if !strings.Contains(err.Error(), ".workflow/f-changelog.md") {
		t.Fatalf("error should name the changelog path, got %v", err)
	}
	os.WriteFile(".workflow/f-changelog.md", []byte("- feat: markdown-first docs\n"), 0644) //nolint:errcheck
	if err := validateDocsOutput("f"); err != nil {
		t.Fatalf("user-facing docs artifact gate should pass with a changelog: %v", err)
	}
	// Without a saved workflow strict orchestration is off, so the full docs
	// step passes on the changelog alone at this layer.
	if err := ValidateArtifacts("f", "docs", nil); err != nil {
		t.Fatalf("docs step should pass with a changelog: %v", err)
	}
}
