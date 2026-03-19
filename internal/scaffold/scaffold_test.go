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
