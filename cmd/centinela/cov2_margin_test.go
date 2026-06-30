package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCov2GitDeliverDefaultRuns exercises the real (unstubbed) gitDeliver seam:
// invoked in a non-git temp dir it executes git and returns its error, covering
// the default function body that production uses.
func TestCov2GitDeliverDefaultRuns(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := gitDeliver("status", "--porcelain"); err == nil {
		t.Fatal("git status in a non-repo should return an error")
	}
}

// TestCov2GhCreatePRDefaultRuns exercises the real ghCreatePR seam: invoked in a
// non-git temp dir it shells out to gh (or fails to exec) and returns an error,
// covering the default body. Either outcome runs the function under test.
func TestCov2GhCreatePRDefaultRuns(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := ghCreatePR("feat", "title", "body.md"); err == nil {
		t.Fatal("gh pr create outside a repo should return an error")
	}
}

// TestCov2SynthesizeWriteDraftError drives runSynthesize's WriteDraft error
// branch: the --out target sits under a regular file (not a directory), so the
// draft write fails.
func TestCov2SynthesizeWriteDraftError(t *testing.T) {
	in := writeInventory(t, ntierInventory)
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(blocker, "PROJECT.md") // parent is a file
	if _, err := runSynth(t, in, out, false); err == nil {
		t.Fatal("expected a WriteDraft error when the --out parent is a file")
	}
}
