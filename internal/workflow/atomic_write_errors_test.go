package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A rename that cannot land must surface, naming the target, and must not leave
// the temp file behind for doctor or `git status` to trip over.
func TestWriteFileAtomicRenameFailureCleansUp(t *testing.T) {
	stateRepo(t)
	// A directory at the target path makes rename(2) fail while every earlier
	// step (create, write, chmod, fsync) succeeds.
	target := filepath.Join(WorkflowDir, "alpha.json")
	if err := os.MkdirAll(filepath.Join(target, "occupied"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := WriteFileAtomic(target, []byte("{}"), stateFileMode)
	if err == nil || !strings.Contains(err.Error(), target) {
		t.Fatalf("a failed rename must surface naming %q, got %v", target, err)
	}
	entries, rerr := os.ReadDir(WorkflowDir)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if len(entries) != 1 {
		t.Fatalf("the temp file must be cleaned up after a failed rename: %v", entries)
	}
}

// Save must surface a write failure rather than reporting success, and must
// leave the previous bytes in place.
func TestSaveSurfacesWriteFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("read-only directory enforcement does not apply to root")
	}
	stateRepo(t)
	if err := Save(New("alpha")); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(FilePath("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(WorkflowDir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(WorkflowDir, 0o755) })

	wf, err := Load("alpha")
	if err != nil {
		t.Fatal(err)
	}
	wf.CurrentStep = "tests"
	if err := Save(wf); err == nil {
		t.Fatal("a save into an unwritable state directory must fail")
	}
	after, err := os.ReadFile(FilePath("alpha"))
	if err != nil || string(after) != string(before) {
		t.Fatalf("the previous state must be untouched, got %q (%v)", after, err)
	}
}

// syncDir is best-effort by contract: a directory it cannot open must never
// turn a save that already succeeded into a failure.
func TestSyncDirIsBestEffort(t *testing.T) {
	stateRepo(t)
	syncDir(filepath.Join(WorkflowDir, "does-not-exist")) // must not panic
}

func TestFirstNonNilReturnsFirstError(t *testing.T) {
	first := errors.New("first")
	if got := firstNonNil(nil, first, errors.New("second")); !errors.Is(got, first) {
		t.Fatalf("firstNonNil = %v, want %v", got, first)
	}
	if got := firstNonNil(nil, nil); got != nil {
		t.Fatalf("firstNonNil = %v, want nil", got)
	}
}
