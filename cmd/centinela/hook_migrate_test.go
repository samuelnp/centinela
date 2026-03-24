package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunHookMigrateOutput(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookMigrate(nil, nil) })
		if out != "" {
			t.Fatalf("expected no output outside centinela project, got %q", out)
		}
	})
	os.WriteFile("PROJECT.md.template", []byte("legacy\n"), 0644) //nolint:errcheck
	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookMigrate(nil, nil) })
		if !strings.Contains(out, "DOC MIGRATION REQUIRED") {
			t.Fatalf("expected migration-needed output, got %q", out)
		}
	})
}
