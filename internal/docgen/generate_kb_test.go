package docgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateWritesKBHTML(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	writeFixture(t)
	os.MkdirAll(KBDir, 0755)                                           //nolint:errcheck
	os.WriteFile(filepath.Join(KBDir, "f.md"), []byte(sampleKB), 0644) //nolint:errcheck

	if err := Generate("docs/project-docs/index.html", "Doc"); err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	idx, err := os.ReadFile(filepath.Join(KBDir, "index.html"))
	if err != nil {
		t.Fatalf("kb index missing: %v", err)
	}
	if !strings.Contains(string(idx), "Knowledge Base") {
		t.Fatal("kb index missing heading")
	}
	page, err := os.ReadFile(filepath.Join(KBDir, "f.html"))
	if err != nil {
		t.Fatalf("kb page missing: %v", err)
	}
	if !strings.Contains(string(page), "What it does") {
		t.Fatal("kb page missing sections")
	}
	main, err := os.ReadFile("docs/project-docs/index.html")
	if err != nil {
		t.Fatalf("main missing: %v", err)
	}
	if !strings.Contains(string(main), `href="kb/index.html"`) {
		t.Fatal("main index missing KB nav link")
	}
}

func TestGenerateFailsOnBrokenKB(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	writeFixture(t)
	os.MkdirAll(KBDir, 0755)                                                //nolint:errcheck
	os.WriteFile(filepath.Join(KBDir, "broken.md"), []byte("# nope"), 0644) //nolint:errcheck
	if err := Generate("docs/project-docs/index.html", "Doc"); err == nil {
		t.Fatal("expected generate to fail on broken kb md")
	}
}
