package roadmap

import (
	"os"
	"path/filepath"
	"testing"
)

// A malformed .git pointer must fall back to the temp dir, never guess a path
// and never drop the lock next to roadmap.json.
func TestGitDirPointerRejectsMalformedContent(t *testing.T) {
	for name, body := range map[string]string{
		"empty":            "",
		"bare prefix":      "gitdir:",
		"prefix and space": "gitdir:   \n",
		"not a pointer":    "ref: refs/heads/main\n",
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, ".git"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, ok := gitDirFor(filepath.Join(dir, RoadmapFile)); ok {
				t.Fatalf("%q must not resolve to a git directory", body)
			}
			if got := stateLockPath(filepath.Join(dir, RoadmapFile)); filepath.Dir(got) != filepath.Clean(os.TempDir()) {
				t.Fatalf("want the temp-dir fallback, got %q", got)
			}
		})
	}
}

// An unreadable .git pointer is a failure to resolve, not a reason to skip
// locking or to write into the working tree.
func TestGitDirPointerUnreadableFallsBack(t *testing.T) {
	dir := t.TempDir()
	pointer := filepath.Join(dir, ".git")
	if err := os.WriteFile(pointer, []byte("gitdir: /somewhere"), 0o200); err != nil {
		t.Fatal(err)
	}
	if _, err := os.ReadFile(pointer); err == nil {
		t.Skip("this filesystem ignores the read bit")
	}
	if _, ok := readGitDirPointer(pointer, dir); ok {
		t.Fatal("an unreadable pointer must not resolve")
	}
	if got := stateLockPath(filepath.Join(dir, RoadmapFile)); filepath.Dir(got) != filepath.Clean(os.TempDir()) {
		t.Fatalf("want the temp-dir fallback, got %q", got)
	}
}

// The walk stops at the filesystem root instead of looping forever.
func TestGitDirForStopsAtTheRoot(t *testing.T) {
	if _, ok := gitDirFor(filepath.Join(t.TempDir(), "deep", "nested", RoadmapFile)); ok {
		t.Fatal("no .git anywhere above must resolve to nothing")
	}
}
