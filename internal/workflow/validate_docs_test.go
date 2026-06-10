package workflow

import (
	"os"
	"strings"
	"testing"
)

func TestValidateDocsOutput(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	if err := validateDocsOutput(""); err == nil {
		t.Fatal("expected missing feature error")
	}
	if err := validateDocsOutput("f"); err == nil || !strings.Contains(err.Error(), "documentation output") {
		t.Fatalf("expected missing docs output error, got %v", err)
	}
	os.MkdirAll("docs/project-docs", 0755)                           //nolint:errcheck
	os.WriteFile("docs/project-docs/index.html", []byte("ok"), 0644) //nolint:errcheck
	if err := ValidateArtifacts("f", "docs", nil); err == nil || !strings.Contains(err.Error(), "knowledge base markdown missing") {
		t.Fatalf("expected KB md missing error, got %v", err)
	}
	os.MkdirAll("docs/project-docs/kb", 0755)                     //nolint:errcheck
	os.WriteFile("docs/project-docs/kb/f.md", []byte("ok"), 0644) //nolint:errcheck
	if err := ValidateArtifacts("f", "docs", nil); err == nil || !strings.Contains(err.Error(), "knowledge base page missing") {
		t.Fatalf("expected KB html missing error, got %v", err)
	}
	os.WriteFile("docs/project-docs/kb/f.html", []byte("<html></html>"), 0644) //nolint:errcheck
	if err := ValidateArtifacts("f", "docs", nil); err != nil {
		t.Fatalf("docs step should pass with KB artifacts: %v", err)
	}
}
