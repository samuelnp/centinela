package main

import (
	"os"
	"testing"
)

func TestRunDocsGenerate(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	writeDocsFixture()
	docsOut = "docs/project-docs/index.html"
	docsTitle = "Doc"
	if err := runDocsGenerate(nil, nil); err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if _, err := os.Stat(docsOut); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
}
