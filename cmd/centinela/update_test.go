package main

import (
	"strings"
	"testing"
)

// Version defaults to the "dev" sentinel under `go test` (no release ldflags),
// so these exercise the command wiring fully offline: the dev path makes no
// network call and writes nothing.

func TestRunUpdateDevBuildIsNoOp(t *testing.T) {
	updateCheck = false
	out := capture(t, func() error { return runUpdate(updateCmd, nil) })
	if !strings.Contains(out, "development build") {
		t.Fatalf("update dev output = %q", out)
	}
}

func TestRunUpdateCheckDevBuild(t *testing.T) {
	updateCheck = true
	defer func() { updateCheck = false }()
	out := capture(t, func() error { return runUpdate(updateCmd, nil) })
	if !strings.Contains(out, "development build") {
		t.Fatalf("--check dev output = %q", out)
	}
}

func TestEmitUpdateNoticeDevIsSilent(t *testing.T) {
	out := capture(t, func() error { emitUpdateNotice(); return nil })
	if out != "" {
		t.Fatalf("dev build should emit no notice, got %q", out)
	}
}

func TestUpdateCommandRegistered(t *testing.T) {
	c, _, err := rootCmd.Find([]string{"update"})
	if err != nil || c.Name() != "update" {
		t.Fatalf("update command not registered: %v", err)
	}
	if c.Flags().Lookup("check") == nil {
		t.Fatal("--check flag missing")
	}
}
