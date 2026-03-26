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

func TestRunMigrateSetupInvalidAgent(t *testing.T) {
	old := setupAgent
	defer func() { setupAgent = old }()
	setupAgent = "bad"
	if err := runMigrateSetup(nil, nil); err == nil {
		t.Fatal("expected setup invalid agent error")
	}
}

func TestRunMigrateSetupNoChangesAndApplyError(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	oldApply, oldAgent := applySetupMigration, setupAgent
	defer func() { applySetupMigration, setupAgent = oldApply, oldAgent }()

	applySetupMigration, setupAgent = true, "opencode"
	if err := runMigrateSetup(nil, nil); err != nil {
		t.Fatal(err)
	}
	applySetupMigration = false
	out := captureStdout(t, func() { _ = runMigrateSetup(nil, nil) })
	if !strings.Contains(out, "already up to date") {
		t.Fatalf("expected no changes output, got %q", out)
	}

	d2 := t.TempDir()
	os.Chdir(d2)                                 //nolint:errcheck
	os.WriteFile(".opencode", []byte("x"), 0644) //nolint:errcheck
	applySetupMigration = true
	if err := runMigrateSetup(nil, nil); err == nil {
		t.Fatal("expected setup apply error")
	}
}
