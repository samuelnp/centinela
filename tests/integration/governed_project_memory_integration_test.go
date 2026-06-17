package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/memory"
)

func memCfg() *config.Config {
	enabled := true
	return &config.Config{Memory: config.MemoryConfig{
		Enabled:          &enabled,
		RecallMaxEntries: 10,
		RecallMaxBytes:   4096,
	}}
}

// countMdFiles counts .md files in the entries dir relative to cwd.
func countMdFiles(t *testing.T) int {
	t.Helper()
	dir := filepath.Join(".workflow", "memory", "entries")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
			n++
		}
	}
	return n
}

// SC-01: tests step capture → lesson entry exists.
func TestMemoryIntegration_CaptureLesson(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.MkdirAll(".workflow", 0755)                                                                        //nolint:errcheck
	os.WriteFile(".workflow/alpha-edge-cases.md", []byte("- timeout retry\n- idempotent writes\n"), 0644) //nolint:errcheck

	memory.Capture("alpha", "tests", memCfg())

	if countMdFiles(t) != 1 {
		t.Fatalf("expected 1 entry file after tests capture, got %d", countMdFiles(t))
	}
}

// SC-05: idempotent — second capture does not duplicate.
func TestMemoryIntegration_CaptureIdempotent(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.MkdirAll(".workflow", 0755)                                                    //nolint:errcheck
	os.WriteFile(".workflow/alpha-edge-cases.md", []byte("- race condition\n"), 0644) //nolint:errcheck

	cfg := memCfg()
	memory.Capture("alpha", "tests", cfg)
	memory.Capture("alpha", "tests", cfg) // repeat

	if countMdFiles(t) != 1 {
		t.Fatalf("expected 1 entry after idempotent capture (SC-05), got %d", countMdFiles(t))
	}
}
