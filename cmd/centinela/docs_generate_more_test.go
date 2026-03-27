package main

import (
	"os"
	"testing"
)

func TestRunDocsGenerateFailsWithoutInputs(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	docsOut = "docs/project-docs/index.html"
	docsTitle = "Doc"
	if err := runDocsGenerate(nil, nil); err == nil {
		t.Fatal("expected missing inputs error")
	}
}
