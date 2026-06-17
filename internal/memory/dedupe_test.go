package memory

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// SC-05: writeIfAbsent is idempotent — second write returns false, no new file.
func TestWriteIfAbsentIdempotence(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	e := newEntry("alpha", "tests", TypeLesson, "- body", ".workflow/alpha-edge-cases.md", []string{"lesson"}, time.Now())
	wrote1, err := writeIfAbsent(e)
	if err != nil {
		t.Fatalf("first write failed: %v", err)
	}
	if !wrote1 {
		t.Fatal("expected first write to return true")
	}
	wrote2, err := writeIfAbsent(e)
	if err != nil {
		t.Fatalf("second write failed: %v", err)
	}
	if wrote2 {
		t.Fatal("second write should return false — entry already exists (SC-05)")
	}
	// Confirm exactly one file written.
	files, _ := os.ReadDir(entriesDir)
	count := 0
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".md" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 entry file, got %d", count)
	}
}

// loadEntries returns empty slice when directory does not exist.
func TestLoadEntriesMissingDir(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	entries := loadEntries()
	if len(entries) != 0 {
		t.Fatalf("expected empty, got %d entries", len(entries))
	}
}

// loadEntries skips malformed files and returns well-formed ones.
func TestLoadEntriesSkipsMalformed(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.MkdirAll(entriesDir, 0o755)                                                      //nolint:errcheck
	os.WriteFile(filepath.Join(entriesDir, "bad.md"), []byte("not frontmatter"), 0o644) //nolint:errcheck

	e := newEntry("f", "tests", TypeLesson, "- valid", "src", []string{}, time.Now())
	writeIfAbsent(e) //nolint:errcheck

	entries := loadEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 valid entry, got %d (malformed should be skipped)", len(entries))
	}
}

// loadEntries skips subdirectories.
func TestLoadEntriesSkipsDirectories(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.MkdirAll(filepath.Join(entriesDir, "subdir.md"), 0o755) //nolint:errcheck

	entries := loadEntries()
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries (directory should be skipped), got %d", len(entries))
	}
}
