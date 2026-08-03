package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteReportReplacesAtomically(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "r.md")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeReport(path, []byte("new")); err != nil {
		t.Fatalf("writeReport: %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "new" {
		t.Fatalf("content = %q", data)
	}
	// The temp file must not survive next to the report.
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("stray temp file left behind: %v", entries)
	}
}

func TestWriteReportFailsWhenDirIsMissing(t *testing.T) {
	if err := writeReport(filepath.Join(t.TempDir(), "nope", "r.md"), []byte("x")); err == nil {
		t.Fatal("want an error when the target directory does not exist")
	}
}

// The stamped report must stay world-readable. writeReport used to hand-roll a
// weaker atomic write with no chmod, so os.CreateTemp's 0600 became the report's
// mode and every `centinela artifact stamp` silently downgraded it.
func TestWriteReportKeepsTheReportReadable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "r.md")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeReport(path, []byte("new")); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Fatalf("stamped report mode = %v, want 0644", info.Mode().Perm())
	}
}
