package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunMigrateDocsNoChangesMessage(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	old := applyDocsMigration
	defer func() { applyDocsMigration = old }()
	applyDocsMigration = true
	if err := runMigrateDocs(nil, nil); err != nil {
		t.Fatal(err)
	}
	applyDocsMigration = false
	out := captureStdout(t, func() { _ = runMigrateDocs(nil, nil) })
	if !strings.Contains(out, "already up to date") {
		t.Fatalf("expected up-to-date output, got %q", out)
	}
}
