package main

import (
	"os"
	"testing"
)

func TestSetupClaudeBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	localOld := localFlag
	defer func() { localFlag = localOld }()
	localFlag = false
	if err := setupClaude(); err != nil {
		t.Fatalf("setupClaude initial error: %v", err)
	}
	if err := setupClaude(); err != nil {
		t.Fatalf("setupClaude idempotent error: %v", err)
	}
}

func TestSetupClaudeErrorWhenDotClaudeIsFile(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                          //nolint:errcheck
	os.Chdir(d)                                //nolint:errcheck
	os.WriteFile(".claude", []byte("x"), 0644) //nolint:errcheck
	if err := setupClaude(); err == nil {
		t.Fatal("expected setupClaude error when .claude is a file")
	}
}
