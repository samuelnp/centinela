package evidence

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

func chdirToTemp(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestWriteAtomicAndRead(t *testing.T) {
	chdirToTemp(t)
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	if err := WriteAtomic("alpha", orchestration.RoleBigThinker, s); err != nil {
		t.Fatal(err)
	}
	got, err := Read("alpha", orchestration.RoleBigThinker)
	if err != nil {
		t.Fatal(err)
	}
	if got.Feature != "alpha" {
		t.Fatalf("feature mismatch: %s", got.Feature)
	}
	matches, _ := filepath.Glob(filepath.Join(workflow.WorkflowDir, "*.tmp"))
	if len(matches) != 0 {
		t.Fatalf("temp file not cleaned: %v", matches)
	}
}

func TestReadMissingReturnsTypedError(t *testing.T) {
	chdirToTemp(t)
	_, err := Read("alpha", orchestration.RoleQASeniorEngineer)
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFound(err) {
		t.Fatalf("expected NotFoundError, got %T: %v", err, err)
	}
	if !strings.Contains(err.Error(), "alpha-qa-senior.json") {
		t.Fatalf("error should name path: %v", err)
	}
}

func TestRepairRemovesOrphanTempFiles(t *testing.T) {
	chdirToTemp(t)
	orphan := TempPathFor("alpha", orchestration.RoleBigThinker)
	if err := os.WriteFile(orphan, []byte("partial"), 0o644); err != nil {
		t.Fatal(err)
	}
	removed, err := Repair("alpha")
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 1 || !strings.HasSuffix(removed[0], ".tmp") {
		t.Fatalf("unexpected repair output: %v", removed)
	}
	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Fatalf("orphan still exists: %v", err)
	}
	again, _ := Repair("alpha")
	if len(again) != 0 {
		t.Fatalf("repair not idempotent: %v", again)
	}
}
