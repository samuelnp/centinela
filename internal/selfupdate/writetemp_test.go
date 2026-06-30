package selfupdate

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTempFile(t *testing.T) *os.File {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "wt")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func TestWriteTempSyncError(t *testing.T) {
	orig := syncFn
	syncFn = func(*os.File) error { return errors.New("sync boom") }
	defer func() { syncFn = orig }()
	err := writeTemp(newTempFile(t), []byte("x"), 0o644)
	var e *Error
	if !errors.As(err, &e) || !strings.Contains(e.Msg, "fsync") {
		t.Fatalf("want fsync error, got %v", err)
	}
}

func TestWriteTempChmodError(t *testing.T) {
	orig := chmodFn
	chmodFn = func(*os.File, os.FileMode) error { return errors.New("chmod boom") }
	defer func() { chmodFn = orig }()
	err := writeTemp(newTempFile(t), []byte("x"), 0o644)
	var e *Error
	if !errors.As(err, &e) || !strings.Contains(e.Msg, "chmod") {
		t.Fatalf("want chmod error, got %v", err)
	}
}

func TestReplaceBinaryWriteTempErrorCleansUp(t *testing.T) {
	orig := writeTempFn
	writeTempFn = func(*os.File, []byte, os.FileMode) error { return errors.New("write boom") }
	defer func() { writeTempFn = orig }()
	dir := t.TempDir()
	target := filepath.Join(dir, "bin")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := replaceBinary(target, []byte("new")); err == nil {
		t.Fatal("want error from writeTemp")
	}
	if got, _ := os.ReadFile(target); string(got) != "old" {
		t.Fatal("target modified despite writeTemp failure")
	}
	entries, _ := os.ReadDir(dir)
	for _, en := range entries {
		if strings.HasPrefix(en.Name(), ".centinela-update-") {
			t.Fatal("temp file not cleaned up")
		}
	}
}
