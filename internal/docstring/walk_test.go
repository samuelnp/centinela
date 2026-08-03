package docstring

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFiles_WalksRootsSkipsVendoredAndSortsResults(t *testing.T) {
	dir := t.TempDir()
	mk := func(rel, body string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("src/b.go", "package a\n")
	mk("src/a.go", "package a\n")
	mk("src/a_test.go", "package a\n")
	mk("src/vendor/v.go", "package v\n")
	mk("src/testdata/t.go", "package t\n")
	mk("other/x.go", "package x\n")

	t.Chdir(dir)
	got, err := Files(Options{Roots: []string{"src"}, IncludeInternal: true})
	if err != nil {
		t.Fatalf("Files: %v", err)
	}
	want := []string{"src/a.go", "src/b.go"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestFiles_MissingRootIsSkippedNotAnError(t *testing.T) {
	t.Chdir(t.TempDir())
	got, err := Files(Options{Roots: []string{"nope"}})
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v, err %v", got, err)
	}
}

func TestFiles_DefaultRootsAreUsedWhenUnset(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "cmd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmd", "m.go"), []byte("package m\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	got, err := Files(Options{IncludeInternal: true})
	if err != nil || len(got) != 1 || got[0] != "cmd/m.go" {
		t.Fatalf("got %v, err %v", got, err)
	}
}
