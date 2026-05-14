package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
	"github.com/samuelnp/centinela/internal/worktree"
)

// This integration exercises the building blocks the cmd/centinela
// `loadActiveWorkflows` hook relies on: cwd-based feature resolution and
// per-worktree workflow state isolation. The hook itself is in `package main`
// (not importable). We validate the contract those helpers expose.

func TestHookResolution_WorkflowInsideWorktreeOnly(t *testing.T) {
	repo := initRepoForWorktrees(t)

	// Provision two parallel features and seed each with its own workflow state.
	wtAlpha, err := worktree.Create(repo, "alpha")
	if err != nil {
		t.Fatalf("Create alpha: %v", err)
	}
	wtBeta, err := worktree.Create(repo, "beta")
	if err != nil {
		t.Fatalf("Create beta: %v", err)
	}

	for _, wt := range []string{wtAlpha, wtBeta} {
		if err := os.MkdirAll(filepath.Join(wt, workflow.WorkflowDir), 0755); err != nil {
			t.Fatalf("mkdir workflow dir: %v", err)
		}
	}

	// When cwd is inside alpha's worktree, detection must yield "alpha".
	feat, root := worktree.DetectFeatureFromCwd(wtAlpha)
	if feat != "alpha" {
		t.Fatalf("expected feature=alpha, got %q (root=%q)", feat, root)
	}

	// Symmetric for beta.
	feat, _ = worktree.DetectFeatureFromCwd(wtBeta)
	if feat != "beta" {
		t.Fatalf("expected feature=beta, got %q", feat)
	}

	// Inside a subdirectory of alpha, detection still returns alpha.
	sub := filepath.Join(wtAlpha, "src", "deep")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	feat, _ = worktree.DetectFeatureFromCwd(sub)
	if feat != "alpha" {
		t.Fatalf("nested cwd: expected alpha, got %q", feat)
	}
}

func TestHookResolution_OutsideWorktreeReturnsEmpty(t *testing.T) {
	repo := initRepoForWorktrees(t)
	// cwd is at repo root — not inside any worktree.
	feat, root := worktree.DetectFeatureFromCwd(repo)
	if feat != "" || root != "" {
		t.Fatalf("expected empty outside worktree, got feature=%q root=%q", feat, root)
	}
}
