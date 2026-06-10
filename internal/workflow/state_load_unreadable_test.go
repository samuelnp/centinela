package workflow

import (
	"os"
	"strings"
	"testing"
)

// Scenario 9: an existing-but-unreadable state file is reported as a read
// failure naming the path, NOT as absence. chmod-based; skip as root, where
// permission bits do not apply.
func TestLoadUnreadableIsNotAbsence(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission bits do not apply when running as root")
	}
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(WorkflowDir, 0755); err != nil {
		t.Fatal(err)
	}
	path := FilePath("locked")
	if err := os.WriteFile(path, []byte(`{"feature":"locked"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0644) })

	_, err := Load("locked")
	if err == nil {
		t.Fatal("expected error for unreadable workflow file")
	}
	msg := err.Error()
	if !strings.Contains(msg, path) {
		t.Fatalf("unreadable error must name the state file path %q, got: %v", path, msg)
	}
	if strings.Contains(msg, "no workflow found") {
		t.Fatalf("unreadable file must not be reported as absence, got: %v", msg)
	}
}
