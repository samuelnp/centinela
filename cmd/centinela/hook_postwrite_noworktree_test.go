package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// TestHookPostwriteOutsideWorktreeIsNoOp pins the behavior when the hook fires
// from a directory that is NOT a worktree (e.g. the repo root). In that case
// DetectFeatureFromCwd returns "" and FormatEvidence skips all files silently
// with no diagnostic. This is the documented behavior for cwd-outside-worktree.
//
// Regression for edge-case: "Postwrite silent no-op outside worktree"
// (.workflow/evidence-cli-edge-cases.md §postwrite formatter outside worktree).
func TestHookPostwriteOutsideWorktreeIsNoOp(t *testing.T) {
	// Set cwd to a plain temp dir (not under .worktrees/<feature>/), so
	// DetectFeatureFromCwd returns ("", nil).
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Place a minified evidence JSON in .workflow/.
	path := filepath.Join(d, workflow.WorkflowDir, "alpha-big-thinker.json")
	body := []byte(minifiedJSON)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}

	// Fire the hook — no active feature detected.
	withStdin(t, `{"tool_input":{"file_path":"`+path+`"}}`, func() {
		if err := runHookPostwrite(nil, nil); err != nil {
			t.Fatal(err)
		}
	})

	// File must be UNCHANGED: no reformat happened (no active feature).
	data, _ := os.ReadFile(path)
	if !bytes.Equal(data, body) {
		t.Fatalf("file should be untouched outside worktree; got:\n%s", data)
	}
}
