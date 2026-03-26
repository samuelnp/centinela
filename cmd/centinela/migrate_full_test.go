package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunMigrateUnknownArg(t *testing.T) {
	if err := runMigrate(nil, []string{"x"}); err == nil {
		t.Fatal("expected unknown arg error")
	}
}

func TestRunMigrateApplyAndNoChanges(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	oldApply, oldAgent := applyFullMigration, fullAgent
	defer func() { applyFullMigration, fullAgent = oldApply, oldAgent }()

	applyFullMigration, fullAgent = false, "both"
	preview := captureStdout(t, func() { _ = runMigrate(nil, nil) })
	if !strings.Contains(preview, "DOCS PREVIEW") || !strings.Contains(preview, "SETUP PREVIEW") {
		t.Fatalf("expected docs+setup preview, got %q", preview)
	}

	applyFullMigration = true
	if err := runMigrate(nil, nil); err != nil {
		t.Fatal(err)
	}

	applyFullMigration = false
	out := captureStdout(t, func() { _ = runMigrate(nil, nil) })
	if !strings.Contains(out, "already up to date") {
		t.Fatalf("expected no changes output, got %q", out)
	}
}

func TestRunMigrateApplySetupError(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.WriteFile(".opencode", []byte("x"), 0644) //nolint:errcheck
	oldApply, oldAgent := applyFullMigration, fullAgent
	defer func() { applyFullMigration, fullAgent = oldApply, oldAgent }()
	applyFullMigration, fullAgent = true, "opencode"
	if err := runMigrate(nil, nil); err == nil {
		t.Fatal("expected full migrate apply error")
	}
}
