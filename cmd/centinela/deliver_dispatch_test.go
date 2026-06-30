package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// TestRunDeliverRejectsBadSlug covers the ValidateFeatureSlug guard.
func TestRunDeliverRejectsBadSlug(t *testing.T) {
	setVia(t, "pr")
	if err := runDeliver(nil, []string{"Bad Slug!"}); err == nil ||
		!strings.Contains(err.Error(), "invalid feature slug") {
		t.Fatalf("bad slug should be rejected, got %v", err)
	}
}

// TestRunDeliverMissingWorkflow covers the workflow.Load error branch.
func TestRunDeliverMissingWorkflow(t *testing.T) {
	deliverRepo(t, true)
	setVia(t, "pr")
	if err := runDeliver(nil, []string{"ghost"}); err == nil ||
		!strings.Contains(err.Error(), "no workflow found") {
		t.Fatalf("missing workflow should surface, got %v", err)
	}
}

// TestRunDeliverPRDispatch reaches runDeliverPR (the final dispatch line) with a
// fully-stubbed push + gh so the happy path executes end to end.
func TestRunDeliverPRDispatch(t *testing.T) {
	deliverRepo(t, true)
	if err := workflow.Save(workflow.New("feat")); err != nil {
		t.Fatal(err)
	}
	cleanPushStub(t)
	stubGH(t, true, "https://x/pull/9", nil)
	setVia(t, "pr")
	out := captureStdout(t, func() {
		if err := runDeliver(nil, []string{"feat"}); err != nil {
			t.Fatalf("pr dispatch should succeed, got %v", err)
		}
	})
	if !strings.Contains(out, "pull/9") {
		t.Fatalf("expected dispatch into runDeliverPR, got %q", out)
	}
}

// TestRunDeliverMergeDispatch covers the via==merge branch: a worktree-mode
// workflow passes the Supports gate and dispatches into runMerge (so the error,
// if any, is NOT the "worktree mode required" refusal).
func TestRunDeliverMergeDispatch(t *testing.T) {
	deliverRepo(t, true)
	wf := workflow.New("feat")
	wf.WorktreePath = "/tmp/wt/feat"
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
	setVia(t, "merge")
	err := runDeliver(nil, []string{"feat"})
	if err != nil && strings.Contains(err.Error(), "worktree mode required") {
		t.Fatalf("merge should dispatch past the Supports gate, got refusal: %v", err)
	}
}
