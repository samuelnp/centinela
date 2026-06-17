package githooks

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstall_MkdirAllFailureSurfaces(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file-as-dir trick is POSIX-specific")
	}
	base := t.TempDir()
	// A regular file where a directory component is expected makes MkdirAll fail.
	blocker := filepath.Join(base, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(filepath.Join(blocker, "hooks")); err == nil {
		t.Fatal("Install must surface a MkdirAll failure")
	}
}

func readPreCommit(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "pre-commit"))
	if err != nil {
		t.Fatalf("read pre-commit: %v", err)
	}
	return string(data)
}

func TestInstall_WritesExecutableHook(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "hooks")
	changed, err := Install(dir)
	if err != nil || !changed {
		t.Fatalf("first install must change, got changed=%v err=%v", changed, err)
	}
	path := filepath.Join(dir, "pre-commit")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("hook not created: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("hook must be executable, mode=%v", info.Mode().Perm())
	}
	body := readPreCommit(t, dir)
	if !strings.Contains(body, BeginMarker) || !strings.Contains(body, "centinela precommit") {
		t.Fatalf("hook missing markers/body: %q", body)
	}
}

func TestInstall_TwiceIsNoOp(t *testing.T) {
	dir := t.TempDir()
	if _, err := Install(dir); err != nil {
		t.Fatal(err)
	}
	first := readPreCommit(t, dir)
	changed, err := Install(dir)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("second install must report changed=false")
	}
	if readPreCommit(t, dir) != first {
		t.Fatal("second install must leave the file byte-identical")
	}
}

func TestInstall_PreservesUserHook(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pre-commit")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho pre-existing-hook\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(dir); err != nil {
		t.Fatal(err)
	}
	body := readPreCommit(t, dir)
	if !strings.Contains(body, "echo pre-existing-hook") {
		t.Fatalf("user hook clobbered: %q", body)
	}
	if !strings.Contains(body, BeginMarker) {
		t.Fatalf("centinela block not appended: %q", body)
	}
}
