package main

import (
	"os"
	"testing"
)

func TestRunStartWorkflowDirConflict(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                             //nolint:errcheck
	os.Chdir(d)                                   //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("x"), 0644) //nolint:errcheck
	os.WriteFile(".workflow", []byte("x"), 0644)  //nolint:errcheck
	if err := runStart(nil, []string{"f"}); err == nil {
		t.Fatal("expected mkdir error when .workflow is a file")
	}
}
