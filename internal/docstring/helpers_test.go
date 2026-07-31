package docstring

import (
	"os"
	"path/filepath"
	"testing"
)

// writeGo writes src into a fresh temp dir and returns its slash path.
func writeGo(t *testing.T, name, src string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return filepath.ToSlash(p)
}

// anyRoot accepts any path, so tests may scan files in a temp dir.
func anyRoot() Options { return Options{Roots: []string{"."}, IncludeInternal: true} }

// scan runs the registered Go scanner and fails the test on error.
func scan(t *testing.T, opts Options, files ...string) Report {
	t.Helper()
	s, ok := For(GoLang)
	if !ok {
		t.Fatal("go scanner not registered")
	}
	rep, err := s.Scan(files, opts)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	return rep
}

// names collects the violation names of a report.
func names(r Report) []string {
	out := make([]string, 0, len(r.Violations))
	for _, v := range r.Violations {
		out = append(out, v.Name)
	}
	return out
}
