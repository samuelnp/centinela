package roadmap

import (
	"os"
	"path/filepath"
	"testing"
)

// MkdirAll fails when the parent exists as a regular file. The caller ignores
// the error, but it must be reported rather than swallowed or panicked on.
func TestSaveSummaryStateParentIsAFile(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".workflow", []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SaveSummaryState(SummaryStatePath(), SummaryState{SessionID: "s"}); err == nil {
		t.Fatal("expected an error when .workflow is a regular file")
	}
}

// The rename step fails when the destination is a non-empty directory. The
// temp file must not be left behind.
func TestSaveSummaryStateRenameFailureCleansUpTemp(t *testing.T) {
	t.Chdir(t.TempDir())
	path := SummaryStatePath()
	if err := os.MkdirAll(filepath.Join(path, "occupied"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := SaveSummaryState(path, SummaryState{SessionID: "s", Digest: "d"}); err == nil {
		t.Fatal("expected a rename error when the destination is a non-empty directory")
	}
	leftovers, _ := filepath.Glob(".workflow/.roadmap-digest-*")
	if len(leftovers) != 0 {
		t.Fatalf("temp file left behind after a failed rename: %v", leftovers)
	}
}

// A closed file makes the write step fail, exercising the write-error branch
// and its close-then-report ordering.
func TestWriteAndRenameWriteFailure(t *testing.T) {
	dir := t.TempDir()
	tmp, err := os.CreateTemp(dir, "probe-*")
	if err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}
	if err := writeAndRename(tmp, filepath.Join(dir, "dest"), []byte("x")); err == nil {
		t.Fatal("expected a write error on an already-closed file")
	}
	if _, err := os.Stat(filepath.Join(dir, "dest")); !os.IsNotExist(err) {
		t.Fatal("destination must not exist after a failed write")
	}
}
