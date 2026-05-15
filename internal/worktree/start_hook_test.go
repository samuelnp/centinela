package worktree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/worktree"
)

func TestMaybeProvision_FlagOff_NoOp(t *testing.T) {
	repo := initRepoForWorktrees(t)
	cfg := &config.Config{}
	cfg.Workflow.UseWorktrees = false
	path, err := worktree.MaybeProvision(repo, "alpha", cfg)
	if err != nil {
		t.Fatalf("MaybeProvision: %v", err)
	}
	if path != "" {
		t.Fatalf("expected empty path with flag off, got %q", path)
	}
	if _, err := os.Stat(filepath.Join(repo, ".worktrees")); !os.IsNotExist(err) {
		t.Fatalf(".worktrees should not exist; err=%v", err)
	}
}

func TestMaybeProvision_NilConfig_NoOp(t *testing.T) {
	repo := initRepoForWorktrees(t)
	path, err := worktree.MaybeProvision(repo, "alpha", nil)
	if err != nil {
		t.Fatalf("MaybeProvision: %v", err)
	}
	if path != "" {
		t.Fatalf("expected empty path with nil config, got %q", path)
	}
}

func TestMaybeProvision_NotAGitRepo_NoOp(t *testing.T) {
	repo := t.TempDir()
	cfg := &config.Config{}
	cfg.Workflow.UseWorktrees = true
	path, err := worktree.MaybeProvision(repo, "alpha", cfg)
	if err != nil {
		t.Fatalf("MaybeProvision must not error in a non-git dir: %v", err)
	}
	if path != "" {
		t.Fatalf("expected empty path outside git repo, got %q", path)
	}
}

func TestMaybeProvision_FlagOn_Provisions(t *testing.T) {
	repo := initRepoForWorktrees(t)
	cfg := &config.Config{}
	cfg.Workflow.UseWorktrees = true
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	path, err := worktree.MaybeProvision(repo, "zeta", cfg)
	if err != nil {
		t.Fatalf("MaybeProvision: %v", err)
	}
	if path == "" {
		t.Fatal("expected a worktree path")
	}
	if _, err := os.Stat(filepath.Join(repo, ".worktrees", "zeta")); err != nil {
		t.Fatalf("worktree dir missing: %v", err)
	}
}
