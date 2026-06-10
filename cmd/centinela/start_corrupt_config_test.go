package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// Scenario 5: starting a feature with a corrupted centinela.toml fails loudly
// (error names centinela.toml) and writes no workflow state file.
func TestRunStartCorruptConfigFailsAndWritesNoState(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("PROJECT.md", []byte("Project Stage: existing\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Unparseable TOML: a bare key with no value.
	if err := os.WriteFile("centinela.toml", []byte("this is = = not toml"), 0644); err != nil {
		t.Fatal(err)
	}

	err := runStart(nil, []string{"newfeat"})
	if err == nil {
		t.Fatal("expected runStart to fail on corrupt centinela.toml")
	}
	if !strings.Contains(err.Error(), "centinela.toml") {
		t.Fatalf("error must name centinela.toml, got: %v", err)
	}
	if _, statErr := os.Stat(workflow.FilePath("newfeat")); statErr == nil {
		t.Fatal("no workflow state file must be created when start fails on config")
	}
}
