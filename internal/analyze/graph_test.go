package analyze

import (
	"os"
	"path/filepath"
	"testing"
)

// chdirModule writes a minimal two-package module and chdirs into it so the
// golist-backed graph runs against a real `go list` invocation.
func chdirModule(t *testing.T, goMod string) {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	mkFile(t, filepath.Join(d, "go.mod"), goMod)
	mkFile(t, filepath.Join(d, "b", "b.go"), "package b\n\nfunc B() {}\n")
	mkFile(t, filepath.Join(d, "a", "a.go"), "package a\n\nimport _ \"fixturemod/b\"\n")
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
}

func TestBuildGraph_GoPackagesEdges(t *testing.T) {
	chdirModule(t, "module fixturemod\n\ngo 1.21\n")
	g := buildGraph([]Manifest{{Kind: "go-mod", Path: "go.mod"}})
	if g.Kind != "go-packages" || g.Module != "fixturemod" {
		t.Fatalf("graph head: %#v", g)
	}
	found := false
	for _, e := range g.Edges {
		if e.From == "a" && e.To == "b" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected module-internal edge a->b, got %#v", g.Edges)
	}
}

func TestBuildGraph_GoListFailureBestEffort(t *testing.T) {
	// Malformed go.mod -> `go list -m` fails -> best-effort empty graph + Note.
	chdirBroken(t)
	g := buildGraph([]Manifest{{Kind: "go-mod", Path: "go.mod"}})
	if g.Kind != "go-packages" || len(g.Edges) != 0 {
		t.Fatalf("go list failure must yield empty go-packages graph: %#v", g)
	}
	if g.Note == "" {
		t.Fatal("go list failure must record a diagnostic Note")
	}
}

func chdirBroken(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	mkFile(t, filepath.Join(d, "go.mod"), "not a valid go.mod\n")
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
}

func TestBuildGraph_DeclaredDepsForNonGo(t *testing.T) {
	g := buildGraph([]Manifest{{Kind: "npm", Path: "package.json", Deps: []string{"react", "next"}}})
	if g.Kind != "declared-deps" {
		t.Fatalf("non-Go manifest deps must be declared-deps: %#v", g)
	}
	if len(g.Edges) != 2 || g.Edges[0].To != "next" || g.Edges[1].To != "react" {
		t.Fatalf("declared edges must be sorted: %#v", g.Edges)
	}
}

func TestBuildGraph_NoneWhenEmpty(t *testing.T) {
	g := buildGraph(nil)
	if g.Kind != "none" || len(g.Edges) != 0 {
		t.Fatalf("no manifests/deps must yield none: %#v", g)
	}
}
