package acceptance_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// loadIndex reads web/index.html relative to the repo root, resolved
// via runtime.Caller so the test is CWD-independent.
func loadIndex(t *testing.T) string {
	t.Helper()
	_, f, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(f), "..", "..")
	data, err := os.ReadFile(filepath.Join(root, "web", "index.html"))
	if err != nil {
		t.Fatalf("loadIndex: cannot read web/index.html: %v", err)
	}
	return string(data)
}
