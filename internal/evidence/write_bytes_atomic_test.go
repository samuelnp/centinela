package evidence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteBytesAtomic_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.json")
	data := []byte(`{"key":"value"}`)
	if err := WriteBytesAtomic(target, data); err != nil {
		t.Fatalf("WriteBytesAtomic: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("content mismatch: got %q want %q", got, data)
	}
	// Temp file must be cleaned up
	if _, err := os.Stat(target + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temp file should not exist after successful write")
	}
}

func TestWriteBytesAtomic_Overwrite(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.json")
	if err := WriteBytesAtomic(target, []byte("first")); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := WriteBytesAtomic(target, []byte("second")); err != nil {
		t.Fatalf("second write: %v", err)
	}
	got, _ := os.ReadFile(target)
	if string(got) != "second" {
		t.Fatalf("overwrite failed: got %q", got)
	}
}

func TestWriteBytesAtomic_BadPath(t *testing.T) {
	err := WriteBytesAtomic("/nonexistent-dir/file.json", []byte("x"))
	if err == nil {
		t.Fatal("expected error writing to invalid path")
	}
}
