package main

import (
	"os"
	"testing"
)

func TestRunHookMigrateNoChangesAfterApply(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.WriteFile("PROJECT.md.template", []byte("x\n"), 0644) //nolint:errcheck
	oldApply, oldAgent := applyFullMigration, fullAgent
	defer func() { applyFullMigration, fullAgent = oldApply, oldAgent }()
	applyFullMigration, fullAgent = true, "both"
	if err := runMigrate(nil, nil); err != nil {
		t.Fatal(err)
	}
	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookMigrate(nil, nil) })
		if out != "" {
			t.Fatalf("expected no migration warning after apply, got %q", out)
		}
	})
}
