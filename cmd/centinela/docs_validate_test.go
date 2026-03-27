package main

import (
	"os"
	"testing"
)

func TestRunDocsValidate(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	if err := runDocsValidate(nil, nil); err == nil {
		t.Fatal("expected missing inputs error")
	}
	writeDocsFixture()
	if err := runDocsValidate(nil, nil); err != nil {
		t.Fatalf("validate should pass: %v", err)
	}
}
