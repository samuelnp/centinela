package workflow

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// ActiveWorkflows must warn (not silently drop) when a state file whose name
// matches its feature is corrupt and fails to Load.
func TestActiveWorkflowsWarnsOnCorruptStateFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(WorkflowDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(FilePath("broken"), []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}

	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	out := ActiveWorkflows(WorkflowDir)
	w.Close() //nolint:errcheck
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r) //nolint:errcheck
	stderr := buf.String()

	if !strings.Contains(stderr, "workflow warning:") {
		t.Fatalf("expected a stderr warning for corrupt state file, got: %q", stderr)
	}
	if len(out) != 0 {
		t.Fatalf("corrupt file must not appear as an active workflow, got %d", len(out))
	}
}
