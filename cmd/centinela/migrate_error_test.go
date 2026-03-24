package main

import (
	"os"
	"testing"
)

func TestRunMigrateDocsApplyError(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.WriteFile("docs", []byte("x"), 0644) //nolint:errcheck
	old := applyDocsMigration
	defer func() { applyDocsMigration = old }()
	applyDocsMigration = true
	if err := runMigrateDocs(nil, nil); err == nil {
		t.Fatal("expected migrate apply error")
	}
}
