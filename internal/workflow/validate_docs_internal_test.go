package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// internalDocsFixture chdirs into a temp repo with an internal (no surface)
// feature brief so validateDocsOutput takes the changelog path.
func internalDocsFixture(t *testing.T, feature string) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/features/"+feature+".md", []byte("# "+feature+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestValidateDocsInternalMissingChangelogNamesFile(t *testing.T) {
	internalDocsFixture(t, "in")
	err := validateDocsOutput("in")
	if err == nil || !strings.Contains(err.Error(), "changelog entry missing") {
		t.Fatalf("missing changelog must be named, got %v", err)
	}
	if !strings.Contains(err.Error(), "in-changelog.md") {
		t.Fatalf("error should name the changelog path, got %v", err)
	}
}

func TestValidateDocsInternalBlankChangelogNamesEmpty(t *testing.T) {
	internalDocsFixture(t, "in")
	path := filepath.Join(WorkflowDir, "in-changelog.md")
	if err := os.WriteFile(path, []byte("   \n\t\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := validateDocsOutput("in")
	if err == nil || !strings.Contains(err.Error(), "changelog entry is empty") {
		t.Fatalf("blank changelog must report empty distinctly, got %v", err)
	}
}

func TestValidateDocsInternalPassesWithOneLinerAndNoKB(t *testing.T) {
	internalDocsFixture(t, "in")
	path := filepath.Join(WorkflowDir, "in-changelog.md")
	if err := os.WriteFile(path, []byte("- refactor: right-size the docs step\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// No KB markdown, no portal index.html — the internal path must still pass.
	if err := validateDocsOutput("in"); err != nil {
		t.Fatalf("internal docs step with a one-line changelog must pass, got %v", err)
	}
}
