package analyze

import (
	"os"
	"path/filepath"
	"testing"
)

// writeF creates rel (and any parent dirs) under dir with a one-byte body, so
// the parent directory counts as populated.
func writeF(t *testing.T, dir, rel string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHasSource(t *testing.T) {
	cases := []struct {
		name      string
		files     []string // files (rel) to create; their parent dirs become populated
		emptyDirs []string // empty dirs to create
		want      bool
	}{
		{"empty dir is greenfield", nil, nil, false},
		{"go.mod manifest", []string{"go.mod"}, nil, true},
		{"Makefile only", []string{"Makefile"}, nil, true},
		{"package.json manifest", []string{"package.json"}, nil, true},
		{"populated src dir", []string{"src/main.go"}, nil, true},
		{"empty src dir is not a signal", nil, []string{"src"}, false},
		{"populated internal dir", []string{"internal/x.go"}, nil, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range c.files {
				writeF(t, dir, f)
			}
			for _, d := range c.emptyDirs {
				if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if got := HasSource(dir); got != c.want {
				t.Fatalf("HasSource(%s)=%v want %v", c.name, got, c.want)
			}
		})
	}
}

func TestDirHasEntry(t *testing.T) {
	dir := t.TempDir()
	if dirHasEntry(filepath.Join(dir, "nope")) {
		t.Fatal("non-existent path must not be a signal")
	}
	file := filepath.Join(dir, "afile")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if dirHasEntry(file) {
		t.Fatal("a plain file (not a dir) must not be a signal")
	}
	empty := filepath.Join(dir, "empty")
	if err := os.MkdirAll(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	if dirHasEntry(empty) {
		t.Fatal("an empty dir must not be a signal")
	}
	writeF(t, empty, "child")
	if !dirHasEntry(empty) {
		t.Fatal("a populated dir must be a signal")
	}
}
