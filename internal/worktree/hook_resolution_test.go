package worktree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
	"github.com/samuelnp/centinela/internal/worktree"
)

func TestHookResolution_WorkflowInsideWorktreeOnly(t *testing.T) {
	repo := initRepoForWorktrees(t)

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

	feat, root := worktree.DetectFeatureFromCwd(wtAlpha)
	if feat != "alpha" {
		t.Fatalf("expected feature=alpha, got %q (root=%q)", feat, root)
	}

	feat, _ = worktree.DetectFeatureFromCwd(wtBeta)
	if feat != "beta" {
		t.Fatalf("expected feature=beta, got %q", feat)
	}

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
	feat, root := worktree.DetectFeatureFromCwd(repo)
	if feat != "" || root != "" {
		t.Fatalf("expected empty outside worktree, got feature=%q root=%q", feat, root)
	}
}
