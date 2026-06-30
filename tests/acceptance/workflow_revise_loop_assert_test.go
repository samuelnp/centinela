// Acceptance: specs/workflow-revise-loop.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"testing"
)

// mustGone asserts each .workflow/<name> evidence file no longer exists.
func mustGone(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, n := range names {
		if _, err := os.Stat(filepath.Join(dir, ".workflow", n)); !os.IsNotExist(err) {
			t.Fatalf("%s must have been invalidated", n)
		}
	}
}

// mustExist asserts each dir-relative path still exists.
func mustExist(t *testing.T, dir string, rels ...string) {
	t.Helper()
	for _, r := range rels {
		if _, err := os.Stat(filepath.Join(dir, r)); err != nil {
			t.Fatalf("%s must survive: %v", r, err)
		}
	}
}
