package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyReturnsErrorWhenParentIsFile(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(filepath.Join(d, "docs"), []byte("x"), 0644) //nolint:errcheck
	plan := Plan{Items: []Item{{Path: "docs/a.md", content: "ok"}}}
	if err := Apply(d, plan); err == nil {
		t.Fatal("expected apply error when parent path is a file")
	}
}
