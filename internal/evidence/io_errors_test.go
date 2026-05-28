package evidence

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestWriteAtomicMkdirFailsOnFileInPlaceOfDir(t *testing.T) {
	d := chdirToTemp(t)
	// Remove .workflow dir, place a regular file with the same name so
	// MkdirAll fails with ENOTDIR.
	if err := os.RemoveAll(filepath.Join(d, workflow.WorkflowDir)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(workflow.WorkflowDir, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	if err := WriteAtomic("alpha", orchestration.RoleBigThinker, s); err == nil {
		t.Fatal("expected mkdir failure")
	}
}

func TestWriteBytesAtomicWriteFailsWhenTargetIsDir(t *testing.T) {
	d := chdirToTemp(t)
	// Create a directory at the target path. Rename(file, dir-existing) on
	// many filesystems errors; the temp open targets <path>.tmp which
	// doesn't conflict, so create a directory at the tmp path instead.
	tmpPath := filepath.Join(d, ".workflow", "foo.txt"+tempSuffix)
	if err := os.MkdirAll(tmpPath, 0o755); err != nil {
		t.Fatal(err)
	}
	err := writeBytesAtomic(filepath.Join(d, ".workflow", "foo.txt"), []byte("x"))
	if err == nil {
		t.Fatal("expected open temp to fail when path is a dir")
	}
}

func TestLockMkdirFailureWhenWorkflowIsFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix-only")
	}
	d := chdirToTemp(t)
	if err := os.RemoveAll(filepath.Join(d, workflow.WorkflowDir)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(workflow.WorkflowDir, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Lock("alpha", orchestration.RoleBigThinker); err == nil {
		t.Fatal("expected mkdir failure")
	}
}

func TestWriteBytesAtomicRenameFailsOnDirTarget(t *testing.T) {
	d := chdirToTemp(t)
	// Place a directory at the target path so Rename fails.
	target := filepath.Join(d, ".workflow", "blocked")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeBytesAtomic(target, []byte("x")); err == nil {
		t.Fatal("expected rename failure")
	}
}

func TestFirstErrReturnsNonNilError(t *testing.T) {
	got := firstErr(nil, os.ErrPermission, nil)
	if got == nil {
		t.Fatal("expected error")
	}
}

func TestCompanionMkdirFailureWhenWorkflowIsFile(t *testing.T) {
	d := chdirToTemp(t)
	if err := os.RemoveAll(filepath.Join(d, workflow.WorkflowDir)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(workflow.WorkflowDir, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteCompanion("alpha", orchestration.RoleBigThinker, "body"); err == nil {
		t.Fatal("expected mkdir failure")
	}
}
