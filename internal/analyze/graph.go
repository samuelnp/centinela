package analyze

import (
	"strings"

	"github.com/samuelnp/centinela/internal/golist"
)

// buildGraph produces the dependency graph for the repo. When a go.mod is among
// the manifests it builds the real Go package graph via golist (module-internal
// edges only); a golist failure degrades to a best-effort empty graph with a
// Note (AC: go list fails). Otherwise it falls back to declared manifest deps,
// or "none" when nothing is declared.
func buildGraph(manifests []Manifest) DependencyGraph {
	if hasManifest(manifests, "go-mod") {
		return goGraph()
	}
	if edges := declaredEdges(manifests); len(edges) > 0 {
		return DependencyGraph{Kind: "declared-deps", Edges: edges}
	}
	return DependencyGraph{Kind: "none", Edges: []Edge{}}
}

// goGraph loads module-internal package edges via golist. Any toolchain error is
// recorded as a Note with an empty edge list (diagnostic, never fatal).
func goGraph() DependencyGraph {
	g := DependencyGraph{Kind: "go-packages", Edges: []Edge{}}
	module, err := golist.ModulePath()
	if err != nil {
		g.Note = "go list -m failed: " + err.Error()
		return g
	}
	g.Module = module
	pkgs, err := golist.Packages()
	if err != nil {
		g.Note = "go list failed: " + err.Error()
		return g
	}
	g.Edges = moduleEdges(pkgs, module)
	return g
}

// moduleEdges flattens golist packages into sorted module-relative import edges,
// dropping stdlib/third-party imports and self edges.
func moduleEdges(pkgs []golist.Pkg, module string) []Edge {
	seen := map[Edge]bool{}
	for _, p := range pkgs {
		from, ok := relPath(p.ImportPath, module)
		if !ok {
			continue
		}
		for _, group := range [][]string{p.Imports, p.TestImports, p.XTestImports} {
			for _, imp := range group {
				to, ok := relPath(imp, module)
				if !ok || to == from {
					continue
				}
				seen[Edge{From: from, To: to}] = true
			}
		}
	}
	return sortedEdges(seen)
}

// relPath converts a full import path to a module-relative path, returning
// (rel, true) for in-module packages and ("", false) otherwise.
func relPath(importPath, module string) (string, bool) {
	if importPath == module {
		return ".", true
	}
	if rest, ok := strings.CutPrefix(importPath, module+"/"); ok {
		return rest, true
	}
	return "", false
}
