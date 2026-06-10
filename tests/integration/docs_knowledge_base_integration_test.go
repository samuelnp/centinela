package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/docgen"
)

const kbBody = `---
feature: f
summary: A plain-language summary for end users.
audience: end-user
status: done
---

## What it does
It generates a per-feature guide.

## When you'd use it
When you want non-tech users to understand the feature.

## How it behaves
- Reads the spec and brief
- Writes plain-language markdown

## Examples
centinela docs generate
`

func TestDocsGenerateProducesKBHTML(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	setupFixture(t)
	os.MkdirAll(docgen.KBDir, 0755)                                         //nolint:errcheck
	os.WriteFile(filepath.Join(docgen.KBDir, "f.md"), []byte(kbBody), 0644) //nolint:errcheck

	if err := docgen.Generate("docs/project-docs/index.html", "Doc"); err != nil {
		t.Fatalf("generate: %v", err)
	}
	page, err := os.ReadFile(filepath.Join(docgen.KBDir, "f.html"))
	if err != nil {
		t.Fatalf("read kb page: %v", err)
	}
	for _, want := range []string{"What it does", "When you'd use it", "How it behaves", "centinela docs generate"} {
		if !strings.Contains(string(page), want) {
			t.Fatalf("kb page missing %q", want)
		}
	}
}

func setupFixture(t *testing.T) {
	t.Helper()
	os.MkdirAll(".workflow", 0755)                                                                                               //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                                                                           //nolint:errcheck
	os.MkdirAll("docs/plans", 0755)                                                                                              //nolint:errcheck
	os.MkdirAll("specs", 0755)                                                                                                   //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("# P"), 0644)                                                                              //nolint:errcheck
	os.WriteFile("ROADMAP.md", []byte("# R"), 0644)                                                                              //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("# F"), 0644)                                                                      //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("# P"), 0644)                                                                         //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: x\n  Scenario: s"), 0644)                                                   //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"name":"P1","features":[{"name":"f"}]}]}`), 0644)                 //nolint:errcheck
	os.WriteFile(".workflow/roadmap-analysis.json", []byte(`{"role":"senior-product-manager","features":[{"name":"f"}]}`), 0644) //nolint:errcheck
	os.WriteFile(".workflow/roadmap-analysis.md", []byte("# A"), 0644)                                                           //nolint:errcheck
}
