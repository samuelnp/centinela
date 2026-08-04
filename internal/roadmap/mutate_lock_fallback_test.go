package roadmap

import (
	"os"
	"path/filepath"
	"testing"
)

// Outside a repository (the supported no-git mutation path) the lock falls back
// to the OS temp dir rather than dropping a file next to roadmap.json.
func TestStateLockPathFallsBackToTempOutsideARepo(t *testing.T) {
	dir := t.TempDir()
	for _, bad := range []string{"", "gitdir: /nonexistent-target"} {
		if bad != "" {
			if err := os.WriteFile(filepath.Join(dir, ".git"), []byte(bad), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		got := stateLockPath(filepath.Join(dir, RoadmapFile))
		if filepath.Dir(got) != filepath.Clean(os.TempDir()) {
			t.Fatalf("want a temp-dir fallback, got %q", got)
		}
	}
}

// Two different checkouts must never share a lock, and one file always maps to
// the same lock however it is addressed.
func TestStateLockPathIsPerRoadmapFile(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	if stateLockPath(filepath.Join(a, RoadmapFile)) == stateLockPath(filepath.Join(b, RoadmapFile)) {
		t.Fatal("two checkouts must not share one lock")
	}
	t.Chdir(a)
	if stateLockPath(RoadmapFile) != stateLockPath(filepath.Join(a, RoadmapFile)) {
		t.Fatal("a relative and absolute address of one file must map to one lock")
	}
}
