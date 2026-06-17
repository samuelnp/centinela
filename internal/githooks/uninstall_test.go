package githooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUninstall_RemovesBlockKeepsUser(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pre-commit")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho pre-existing-hook\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(dir); err != nil {
		t.Fatal(err)
	}
	changed, err := Uninstall(dir)
	if err != nil || !changed {
		t.Fatalf("uninstall must remove the block, changed=%v err=%v", changed, err)
	}
	body := readPreCommit(t, dir)
	if strings.Contains(body, BeginMarker) {
		t.Fatalf("block not removed: %q", body)
	}
	if !strings.Contains(body, "echo pre-existing-hook") {
		t.Fatalf("user line lost on uninstall: %q", body)
	}
}

func TestUninstall_DeletesCentinelaOnlyHook(t *testing.T) {
	dir := t.TempDir()
	if _, err := Install(dir); err != nil {
		t.Fatal(err)
	}
	changed, err := Uninstall(dir)
	if err != nil || !changed {
		t.Fatalf("uninstall must report changed, got changed=%v err=%v", changed, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "pre-commit")); !os.IsNotExist(err) {
		t.Fatalf("centinela-only hook must be deleted, stat err=%v", err)
	}
}

func TestUninstall_MissingFileIsNoOp(t *testing.T) {
	changed, err := Uninstall(t.TempDir())
	if err != nil || changed {
		t.Fatalf("uninstall of a missing hook must be a no-op, changed=%v err=%v", changed, err)
	}
}

func TestUninstall_ReadErrorSurfaces(t *testing.T) {
	dir := t.TempDir()
	// A directory at the hook path makes os.ReadFile return a non-IsNotExist error.
	if err := os.MkdirAll(filepath.Join(dir, "pre-commit"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Uninstall(dir); err == nil {
		t.Fatal("Uninstall must surface a non-IsNotExist read error")
	}
}

func TestReadHookAndIsEmptyHook(t *testing.T) {
	if readHook(filepath.Join(t.TempDir(), "nope")) != "" {
		t.Fatal("readHook of a missing file must return empty string")
	}
	if !isEmptyHook("#!/bin/sh\n   \n") {
		t.Fatal("bare shebang + whitespace must count as empty")
	}
	if isEmptyHook("echo hi\n") {
		t.Fatal("a real command line must not count as empty")
	}
}
