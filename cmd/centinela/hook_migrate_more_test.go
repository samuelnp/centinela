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
	old := applyDocsMigration
	defer func() { applyDocsMigration = old }()
	applyDocsMigration = true
	if err := runMigrateDocs(nil, nil); err != nil {
		t.Fatal(err)
	}
	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookMigrate(nil, nil) })
		if out != "" {
			t.Fatalf("expected no migration warning after apply, got %q", out)
		}
	})
}
