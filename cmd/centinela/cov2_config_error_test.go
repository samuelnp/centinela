package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// writeCorruptConfig chdirs into a temp dir holding an unparseable
// centinela.toml so config.Load() returns a non-nil error.
func writeCorruptConfig(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "centinela.toml"), []byte("x = = not toml"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d)
	return d
}

func TestCov2PrGateConfigLoadError(t *testing.T) {
	writeCorruptConfig(t)
	if err := runPrGate(nil, nil); err == nil || !strings.Contains(err.Error(), "centinela.toml") {
		t.Fatalf("pr-gate must surface corrupt config, got %v", err)
	}
}

func TestCov2PrecommitConfigLoadError(t *testing.T) {
	writeCorruptConfig(t)
	if err := runPrecommit(nil, nil); err == nil || !strings.Contains(err.Error(), "centinela.toml") {
		t.Fatalf("precommit must surface corrupt config, got %v", err)
	}
}

func TestCov2MigrateConfigLoadError(t *testing.T) {
	writeCorruptConfig(t)
	prev := fullAgent
	fullAgent = "both"
	t.Cleanup(func() { fullAgent = prev })
	if err := runMigrate(nil, nil); err == nil || !strings.Contains(err.Error(), "centinela.toml") {
		t.Fatalf("migrate must surface corrupt config, got %v", err)
	}
}

// TestCov2BuildStatusLineViewConfigFallback drives the silent config fallback
// (cfg = &Config{}) inside buildStatusLineView when centinela.toml is corrupt.
func TestCov2BuildStatusLineViewConfigFallback(t *testing.T) {
	writeCorruptConfig(t)
	wf := workflow.New("alpha")
	view := buildStatusLineView([]*workflow.Workflow{wf})
	if len(view.Primary) == 0 || view.Primary[0] != "WF:alpha" {
		t.Fatalf("expected a primary status line for alpha, got %+v", view.Primary)
	}
}
