package evidence

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// TestRepairDeletesLiveTempFile documents current behavior: Repair removes any
// .json.tmp file matching the feature prefix regardless of whether a concurrent
// writer just created it. There is no mtime guard. This test pins the observed
// race window so a future implementor adding an mtime guard notices it.
//
// Regression for edge-case: "Repair sweeps live temp files during concurrent append"
// (.workflow/evidence-cli-edge-cases.md §Repair has no mtime guard).
func TestRepairDeletesLiveTempFile(t *testing.T) {
	chdirToTemp(t)

	// Simulate a writer that created the temp file but has not yet renamed it.
	tmpPath := TempPathFor("alpha", orchestration.RoleBigThinker)
	if err := os.WriteFile(tmpPath, []byte(`{"partial": true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Repair runs between writeTempFile and os.Rename in the concurrent writer.
	removed, err := Repair("alpha")
	if err != nil {
		t.Fatalf("repair: %v", err)
	}

	// Current behavior: the live temp file IS deleted. This is the documented
	// limitation — no mtime guard exists. The test name makes this explicit so
	// a future PR adding mtime protection knows to update this assertion.
	if len(removed) != 1 {
		t.Fatalf("expected Repair to remove the live temp file (no mtime guard); got %v", removed)
	}
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Fatal("expected temp file to be gone after Repair")
	}
}

// TestRepairDoesNotRemoveLockFiles documents that Repair only removes .json.tmp
// orphans and does NOT clean up .lock files. Lock files accumulate silently
// after crashes but are otherwise harmless (OS releases flock on fd close).
//
// Regression for edge-case: "Lock file cleanup" (.workflow/evidence-cli-edge-cases.md).
func TestRepairDoesNotRemoveLockFiles(t *testing.T) {
	chdirToTemp(t)

	// Simulate a crashed process that left its .lock file behind.
	lockFile := lockPath("alpha", orchestration.RoleBigThinker)
	if err := os.WriteFile(lockFile, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	removed, err := Repair("alpha")
	if err != nil {
		t.Fatalf("repair: %v", err)
	}

	// Repair must report nothing removed — it does not touch .lock files.
	if len(removed) != 0 {
		t.Fatalf("Repair should not remove .lock files, removed: %v", removed)
	}
	if _, err := os.Stat(lockFile); err != nil {
		t.Fatalf("lock file should still exist after Repair: %v", err)
	}
}
