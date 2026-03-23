package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractCreatesAndSkips(t *testing.T) {
	d := t.TempDir()
	res1, err := Extract(d)
	if err != nil {
		t.Fatalf("Extract #1: %v", err)
	}
	if len(res1.Created) == 0 {
		t.Fatal("expected created files")
	}
	if _, err := os.Stat(filepath.Join(d, "CLAUDE.md")); err != nil {
		t.Fatalf("missing CLAUDE.md: %v", err)
	}
	res2, err := Extract(d)
	if err != nil {
		t.Fatalf("Extract #2: %v", err)
	}
	if len(res2.Skipped) == 0 {
		t.Fatal("expected skipped files on second run")
	}
}

func TestExtractErrorOnInvalidTarget(t *testing.T) {
	d := t.TempDir()
	bad := filepath.Join(d, "bad")
	os.WriteFile(bad, []byte("x"), 0644) //nolint:errcheck
	if _, err := Extract(filepath.Join(bad, "nested")); err == nil {
		t.Fatal("expected extract error for invalid target path")
	}
}
