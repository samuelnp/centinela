package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stateRepo chdirs into an empty repo with a .workflow/ directory.
func stateRepo(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	t.Chdir(d)
	if err := os.MkdirAll(WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestWriteFileAtomicReplacesAndKeepsMode(t *testing.T) {
	stateRepo(t)
	target := filepath.Join(WorkflowDir, "alpha.json")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(target, []byte("new"), stateFileMode); err != nil {
		t.Fatalf("WriteFileAtomic: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil || string(got) != "new" {
		t.Fatalf("content = %q, err = %v", got, err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Fatalf("mode = %v, want 0644 (os.CreateTemp opens at 0600)", info.Mode().Perm())
	}
}

// TestWriteFileAtomicLeavesNoTemp pins the naming contract: after a successful
// write nothing is left beside the target, and the temp name the writer picks
// matches neither the ActiveWorkflows glob nor the doctor evidence-orphan glob.
func TestWriteFileAtomicLeavesNoTemp(t *testing.T) {
	stateRepo(t)
	target := filepath.Join(WorkflowDir, "alpha.json")
	if err := WriteFileAtomic(target, []byte("{}"), stateFileMode); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(WorkflowDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "alpha.json" {
		t.Fatalf("temp file left behind: %v", entries)
	}
	tmp, err := writeTempSibling(WorkflowDir, "alpha.json", []byte("x"), stateFileMode)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(tmp) })
	assertNoGlobMatch(t, "*.json.tmp", tmp) // doctor's orphan sweep
	assertNoGlobMatch(t, "*.json", tmp)     // ActiveWorkflows' state-file scan
}

func assertNoGlobMatch(t *testing.T, pattern, tmp string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(WorkflowDir, pattern))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range matches {
		if m == tmp {
			t.Fatalf("temp %q matches %q: doctor would go red with a repair that removes nothing", tmp, pattern)
		}
	}
}

// TestWriteFileAtomicErrorNamesTarget: the operator asked to write the state
// file, so that is the path the failure must name — not the temp.
func TestWriteFileAtomicErrorNamesTarget(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("read-only directory enforcement does not apply to root")
	}
	stateRepo(t)
	if err := os.Chmod(WorkflowDir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(WorkflowDir, 0o755) })
	target := filepath.Join(WorkflowDir, "alpha.json")
	err := WriteFileAtomic(target, []byte("{}"), stateFileMode)
	if err == nil || !strings.Contains(err.Error(), target) {
		t.Fatalf("error must name %q, got %v", target, err)
	}
	if _, statErr := os.Stat(target); statErr == nil {
		t.Fatal("a failed write must not leave a partial state file")
	}
}
