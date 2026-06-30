package analyze

import (
	"testing"

	"github.com/samuelnp/centinela/internal/golist"
)

// TestModuleEdges_FiltersSortsAndSelf drives moduleEdges with synthetic golist
// packages so every relPath branch (module root ".", in-module, out-of-module),
// the stdlib/self drop, and both sortedEdges comparator branches are hit.
func TestModuleEdges_FiltersSortsAndSelf(t *testing.T) {
	pkgs := []golist.Pkg{
		{ImportPath: "mod", Imports: []string{"mod/a"}},                            // from "."
		{ImportPath: "mod/a", Imports: []string{"mod/b", "mod/c", "fmt", "mod/a"}}, // stdlib + self dropped
		{ImportPath: "other/x", Imports: []string{"mod/a"}},                        // from not in module -> skipped
	}
	edges := moduleEdges(pkgs, "mod")
	want := map[string]bool{".|a": true, "a|b": true, "a|c": true}
	if len(edges) != len(want) {
		t.Fatalf("expected %d edges, got %#v", len(want), edges)
	}
	for i, e := range edges {
		if !want[e.From+"|"+e.To] {
			t.Fatalf("unexpected edge %#v", e)
		}
		if i > 0 {
			prev := edges[i-1]
			if prev.From > e.From || (prev.From == e.From && prev.To > e.To) {
				t.Fatalf("edges not sorted: %#v", edges)
			}
		}
	}
}

func TestRelPath_OutOfModule(t *testing.T) {
	if _, ok := relPath("github.com/other/pkg", "mod"); ok {
		t.Fatal("out-of-module import must not be relativized")
	}
}
