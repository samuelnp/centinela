package worktree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// WritePending must surface an error when .workflow cannot be created
// (a regular file already occupies that name).
func TestWritePending_WorkflowDirBlockedErrors(t *testing.T) {
	d := t.TempDir()
	_ = os.WriteFile(filepath.Join(d, ".workflow"), []byte("x"), 0o644)
	err := worktree.WritePending(d,
		worktree.MergeOutcome{Feature: "alpha", TextConflict: true})
	if err == nil {
		t.Fatal("WritePending must error when .workflow is not a directory")
	}
}

// LoadPending must surface a read error distinct from absence when the
// marker path is a directory rather than a file.
func TestLoadPending_PathIsDirectoryErrors(t *testing.T) {
	d := t.TempDir()
	_ = os.MkdirAll(worktree.PendingPath(d, "beta"), 0o755)
	m, err := worktree.LoadPending(d, "beta")
	if err == nil {
		t.Fatal("LoadPending must error when marker path is a directory")
	}
	if m != nil {
		t.Fatalf("LoadPending must not return a marker on read error: %+v", m)
	}
}

// ClearPending must surface a non-absence removal error (path is a
// non-empty directory, which os.Remove cannot delete).
func TestClearPending_RemovalErrorSurfaced(t *testing.T) {
	d := t.TempDir()
	mp := worktree.PendingPath(d, "gamma")
	_ = os.MkdirAll(filepath.Join(mp, "child"), 0o755)
	if err := worktree.ClearPending(d, "gamma"); err == nil {
		t.Fatal("ClearPending must surface a non-absence removal error")
	}
}

// ResolveMerge must surface the isDirty git error when run outside a git
// repository (after a marker is present so it reaches the dirty check).
func TestResolveMerge_IsDirtyErrorSurfaced(t *testing.T) {
	d := t.TempDir() // not a git repo
	if err := worktree.WritePending(d,
		worktree.MergeOutcome{Feature: "delta", TextConflict: true}); err != nil {
		t.Fatalf("WritePending: %v", err)
	}
	_, err := worktree.ResolveMerge(d, "delta",
		func(string) (string, error) { return "complete", nil })
	if err == nil {
		t.Fatal("ResolveMerge must surface git status error outside a repo")
	}
}
