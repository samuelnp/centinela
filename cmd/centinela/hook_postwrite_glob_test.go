package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// TestHookPostwriteGlobSkipsNonWorkflowJSON covers the `continue` branch in
// runHookPostwrite where `.workflow/*.json` enumerates files that are not
// the workflow state file (e.g. evidence JSON dropped by the agent).
func TestHookPostwriteGlobSkipsNonWorkflowJSON(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Drop a JSON file whose base name does NOT correspond to any workflow
	// state file — the loop must `continue` past it.
	path := filepath.Join(workflow.WorkflowDir, "alpha-big-thinker.json")
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	withStdin(t, "{}", func() {
		if err := runHookPostwrite(nil, nil); err != nil {
			t.Fatal(err)
		}
	})
}
