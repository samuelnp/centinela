package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunMigrateDocsPreviewAndApply(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	old := applyDocsMigration
	defer func() { applyDocsMigration = old }()
	applyDocsMigration = false
	preview := captureStdout(t, func() { _ = runMigrateDocs(nil, nil) })
	if !strings.Contains(preview, "Managed docs requiring migration") {
		t.Fatalf("expected migration preview output, got %q", preview)
	}
	applyDocsMigration = true
	if err := runMigrateDocs(nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat("CLAUDE.md"); err != nil {
		t.Fatalf("expected migration apply to create CLAUDE.md: %v", err)
	}
}
