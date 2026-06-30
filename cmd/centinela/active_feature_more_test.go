package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// TestActiveWorkflowWorktreeMode exercises the worktree branch: when cwd is
// inside a `.worktrees/<feature>` path and a matching workflow file is present,
// activeWorkflow resolves that feature directly (not the root-scan fallback).
func TestActiveWorkflowWorktreeMode(t *testing.T) {
	root := t.TempDir()
	wt := filepath.Join(root, ".worktrees", "wtfeat")
	if err := os.MkdirAll(filepath.Join(wt, workflow.WorkflowDir), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(wt)
	if err := workflow.Save(workflow.New("wtfeat")); err != nil {
		t.Fatal(err)
	}
	wf := activeWorkflow(mustGetwd())
	if wf == nil || wf.Feature != "wtfeat" {
		t.Fatalf("expected wtfeat from worktree path, got %+v", wf)
	}
}

// TestActiveWorkflowWorktreePathButNoState falls through to the root scan when
// the cwd looks like a worktree but no state file exists for that feature.
func TestActiveWorkflowWorktreePathButNoState(t *testing.T) {
	root := t.TempDir()
	wt := filepath.Join(root, ".worktrees", "ghostfeat")
	if err := os.MkdirAll(filepath.Join(wt, workflow.WorkflowDir), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(wt)
	if wf := activeWorkflow(mustGetwd()); wf != nil {
		t.Fatalf("missing worktree state should yield no workflow, got %+v", wf)
	}
}
