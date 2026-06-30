package selfupdate

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteTempClosedFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "x")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	werr := writeTemp(f, []byte("data"), 0o644)
	var e *Error
	if !errors.As(werr, &e) || e.Kind != KindReplace {
		t.Fatalf("want replace error on closed file, got %v", werr)
	}
}

func TestReplaceBinaryRenameError(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "sub")
	if err := os.MkdirAll(filepath.Join(target, "child"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := replaceBinary(target, []byte("x")) // rename over a non-empty dir fails
	var e *Error
	if !errors.As(err, &e) || e.Kind != KindReplace {
		t.Fatalf("want replace error, got %v", err)
	}
	entries, _ := os.ReadDir(dir)
	for _, en := range entries {
		if strings.HasPrefix(en.Name(), ".centinela-update-") {
			t.Fatal("temp file not cleaned after rename failure")
		}
	}
}
