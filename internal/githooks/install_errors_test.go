package githooks

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestInstallWriteError: when pre-commit is itself a directory, WriteFile fails
// and Install surfaces the error rather than panicking.
func TestInstallWriteError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory-as-file semantics differ on Windows")
	}
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "pre-commit"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(dir); err == nil {
		t.Fatal("expected Install to fail when pre-commit is a directory")
	}
}

// TestUninstallReadError: a pre-commit that is a directory yields a non-NotExist
// read error, surfaced by Uninstall (not treated as a no-op).
func TestUninstallReadError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory-as-file semantics differ on Windows")
	}
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "pre-commit"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Uninstall(dir); err == nil {
		t.Fatal("expected Uninstall to surface the read error")
	}
}

// TestUninstallNoBlock: a hook with no centinela block is left untouched
// (changed=false), not rewritten or deleted.
func TestUninstallNoBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pre-commit")
	body := "#!/bin/sh\necho user\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	changed, err := Uninstall(dir)
	if err != nil || changed {
		t.Fatalf("expected no-op uninstall, got changed=%v err=%v", changed, err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != body {
		t.Fatalf("hook altered: %q", got)
	}
}

// TestUninstallDeletesBareShebangRemnant: when removing the block leaves only a
// shebang, the file is deleted.
func TestUninstallDeletesBareShebangRemnant(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pre-commit")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+Block), 0o755); err != nil {
		t.Fatal(err)
	}
	changed, err := Uninstall(dir)
	if err != nil || !changed {
		t.Fatalf("expected deletion, got changed=%v err=%v", changed, err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected hook file removed, stat err=%v", err)
	}
}
