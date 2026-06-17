package workflow

import (
	"os"
	"strings"
	"testing"
)

// Scenario 8: an invalid-JSON state file reports the path AND the parse cause,
// and is never mistaken for absence.
func TestLoadCorruptReportsPathAndCause(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(WorkflowDir, 0755); err != nil {
		t.Fatal(err)
	}
	path := FilePath("broken")
	if err := os.WriteFile(path, []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load("broken")
	if err == nil {
		t.Fatal("expected error for corrupt workflow file")
	}
	msg := err.Error()
	if !strings.Contains(msg, path) {
		t.Fatalf("corrupt error must name the state file path %q, got: %v", path, msg)
	}
	if strings.Contains(msg, "no workflow found") {
		t.Fatalf("corrupt file must not be reported as absence, got: %v", msg)
	}
	// The underlying parse failure must be wrapped (json error mentions a
	// character offset / invalid syntax).
	if !strings.Contains(msg, "invalid workflow file") {
		t.Fatalf("corrupt error should wrap the parse cause, got: %v", msg)
	}
}
