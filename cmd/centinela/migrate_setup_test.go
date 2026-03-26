package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunMigrateSetupPreviewAndApply(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	oldApply, oldAgent := applySetupMigration, setupAgent
	defer func() { applySetupMigration, setupAgent = oldApply, oldAgent }()
	applySetupMigration, setupAgent = false, "opencode"
	preview := captureStdout(t, func() { _ = runMigrateSetup(nil, nil) })
	if !strings.Contains(preview, "Managed setup assets requiring migration") {
		t.Fatalf("expected setup migration preview output, got %q", preview)
	}
	applySetupMigration = true
	if err := runMigrateSetup(nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(".opencode/plugins/centinela.js"); err != nil {
		t.Fatalf("expected setup apply to create plugin: %v", err)
	}
}

func TestRunMigrateInvalidAgent(t *testing.T) {
	old := fullAgent
	defer func() { fullAgent = old }()
	fullAgent = "nope"
	if err := runMigrate(nil, nil); err == nil {
		t.Fatal("expected invalid agent error")
	}
}
