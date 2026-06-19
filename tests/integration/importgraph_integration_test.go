package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/importgraph"
)

// TestImportGraph_GoProviderRealModule exercises the go backend end-to-end
// against a real two-package module via the actual `go list` toolchain.
func TestImportGraph_GoProviderRealModule(t *testing.T) {
	dir := t.TempDir()
	mk(t, dir, "go.mod", "module fixturemod\n\ngo 1.21\n")
	mk(t, dir, "b/b.go", "package b\n\nfunc B() {}\n")
	mk(t, dir, "a/a.go", "package a\n\nimport _ \"fixturemod/b\"\n")
	chdir(t, dir)

	p, err := importgraph.Select(".", "", "fixturemod", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "go" {
		t.Fatalf("expected auto-selected go provider, got %q", p.Name())
	}
	g, err := p.Load(".")
	if err != nil {
		t.Fatal(err)
	}
	if !hasEdge(g, "a", "b") {
		t.Fatalf("expected scoped edge a -> b, got %+v", g.Pkgs)
	}
}

// TestImportGraph_PythonProviderRealWalker runs the embedded AST walker against
// a real Python package; skipped when python3 is absent so CI stays green.
func TestImportGraph_PythonProviderRealWalker(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	dir := t.TempDir()
	mk(t, dir, filepath.Join("a", "__init__.py"), "import b\n")
	mk(t, dir, filepath.Join("b", "__init__.py"), "x = 1\n")

	p, err := importgraph.Select(dir, "python", "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	g, err := p.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !hasEdge(g, "a", "b") {
		t.Fatalf("python walker should find a -> b, got %+v", g.Pkgs)
	}
}

func hasEdge(g importgraph.Graph, from, to string) bool {
	for _, p := range g.Pkgs {
		if p.Path != from {
			continue
		}
		for _, imp := range p.Imports {
			if imp == to {
				return true
			}
		}
	}
	return false
}

func mk(t *testing.T, dir, rel, body string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
}
