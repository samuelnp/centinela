package docgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleKB = `---
feature: alpha
summary: Plain summary.
audience: end-user
status: done
---

## What it does
It does the thing.

## When you'd use it
When you need the thing.

## How it behaves
- It behaves first
- Then it behaves second

## Examples
Try centinela start alpha.
`

func TestParseKBFile_Valid(t *testing.T) {
	p, err := parseKBFile("docs/project-docs/kb/alpha.md", sampleKB)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if p.Feature != "alpha" || p.Summary != "Plain summary." || p.Status != "done" {
		t.Fatalf("unexpected frontmatter: %#v", p)
	}
	if !strings.Contains(p.WhatItDoes, "does the thing") {
		t.Fatalf("WhatItDoes missing: %q", p.WhatItDoes)
	}
	if !strings.Contains(p.HowItBehaves, "behaves first") || !strings.Contains(p.Examples, "centinela start alpha") {
		t.Fatalf("section bodies wrong: %#v", p)
	}
}

func TestParseKBFile_MissingSection(t *testing.T) {
	bad := strings.Replace(sampleKB, "## What it does\nIt does the thing.\n\n", "", 1)
	_, err := parseKBFile("docs/project-docs/kb/alpha.md", bad)
	if err == nil || !strings.Contains(err.Error(), "What it does") {
		t.Fatalf("expected missing-section error, got %v", err)
	}
}

func TestParseKBFile_UnterminatedFrontmatter(t *testing.T) {
	bad := "---\nfeature: x\n## What it does\nx\n"
	_, err := parseKBFile("docs/project-docs/kb/x.md", bad)
	if err == nil || !strings.Contains(err.Error(), "unterminated") {
		t.Fatalf("expected unterminated error, got %v", err)
	}
}

func TestLoadKBPages_DirAndError(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	os.MkdirAll(KBDir, 0755)                                              //nolint:errcheck
	os.WriteFile(filepath.Join(KBDir, "alpha.md"), []byte(sampleKB), 0644) //nolint:errcheck
	pages, err := loadKBPages()
	if err != nil || len(pages) != 1 {
		t.Fatalf("expected one page, got %d, err %v", len(pages), err)
	}
	os.WriteFile(filepath.Join(KBDir, "broken.md"), []byte("# nope"), 0644) //nolint:errcheck
	if _, err := loadKBPages(); err == nil {
		t.Fatal("expected error for broken kb md")
	}
}
