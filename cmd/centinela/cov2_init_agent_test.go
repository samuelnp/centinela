package main

import (
	"os"
	"strings"
	"testing"
)

// TestCov2SetupOpenCodeConfigWriteError drives setupOpenCode's first error
// branch: opencode.json present as a directory makes the config build/read fail.
func TestCov2SetupOpenCodeConfigWriteError(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir("opencode.json", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := setupOpenCode(); err == nil || !strings.Contains(err.Error(), "opencode.json") {
		t.Fatalf("expected an opencode.json failure, got %v", err)
	}
}

// TestCov2SetupOpenCodePluginWriteError drives the plugin-write error branch:
// the config injects cleanly, then .opencode (present as a file) blocks the
// plugin directory creation.
func TestCov2SetupOpenCodePluginWriteError(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".opencode", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := setupOpenCode(); err == nil || !strings.Contains(err.Error(), ".opencode/plugins") {
		t.Fatalf("expected an OpenCode asset write failure, got %v", err)
	}
}

// TestCov2SetupAiderApplyError drives setupAider's error path by planting a
// directory where a managed Aider file must be written.
func TestCov2SetupAiderApplyError(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir("AGENTS.md", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := setupAider(); err == nil {
		t.Fatal("expected an Aider asset write failure")
	}
}
