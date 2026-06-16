package audit

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadOnDirectoryErrors surfaces a non-IsNotExist read error (exists=false,
// err!=nil) when the path is a directory rather than a file.
func TestLoadOnDirectoryErrors(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "asdir")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(dir); err == nil {
		t.Fatal("reading a directory should error")
	}
}

// TestSaveBadPathErrors fails when the parent path is a file, not a directory.
func TestSaveBadPathErrors(t *testing.T) {
	f := filepath.Join(t.TempDir(), "afile")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Save(filepath.Join(f, "baseline.json"), sampleBaseline()); err == nil {
		t.Fatal("save under a file path should error")
	}
}

// TestSortByHashEmpty is a no-op on an empty slice (boundary).
func TestSortByHashEmpty(t *testing.T) {
	sortByHash(nil)
}
