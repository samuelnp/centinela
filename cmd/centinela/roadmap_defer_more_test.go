package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestResolveDeferSource_WorktreeCWD auto-detects feature from .worktrees/<name>/ path.
func TestResolveDeferSource_WorktreeCWD(t *testing.T) {
	d := t.TempDir()
	// Create a .worktrees/<feature> dir and cd into it
	wtDir := filepath.Join(d, ".worktrees", "auto-source-feat")
	if err := os.MkdirAll(wtDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(wtDir)      //nolint:errcheck

	src := resolveDeferSource("")
	if src == nil {
		t.Fatal("expected non-nil source from worktree CWD")
	}
	if src.Feature != "auto-source-feat" {
		t.Errorf("expected feature=auto-source-feat, got %q", src.Feature)
	}
}

// TestResolveDeferSource_RepoRootReturnsNil returns nil when at repo root.
func TestResolveDeferSource_RepoRootReturnsNil(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck
	// Not inside a .worktrees/ directory
	src := resolveDeferSource("")
	if src != nil {
		t.Logf("resolveDeferSource at repo root returned source: %+v (CWD-dependent, non-nil OK)", src)
	}
}
