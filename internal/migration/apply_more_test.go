package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyWriteFileError(t *testing.T) {
	// The target path itself is an existing directory: MkdirAll on the parent
	// succeeds but WriteFile cannot write over a directory, exercising the
	// WriteFile error branch (distinct from the parent-is-file MkdirAll error).
	d := t.TempDir()
	os.MkdirAll(filepath.Join(d, "x"), 0755) //nolint:errcheck
	plan := Plan{Items: []Item{{Path: "x", content: "data"}}}
	if err := Apply(d, plan); err == nil {
		t.Fatal("expected WriteFile error when target path is a directory")
	}
}
