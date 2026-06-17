package gates

import (
	"fmt"
	"sort"
	"strings"
)

// pkg is a single module package and its intra-module imports, already scoped
// to the module (paths are import paths relative to the module root, e.g.
// "internal/config"). Built by the loader; consumed by pure check logic.
type pkg struct {
	Path    string
	Imports []string
}

// checkEdges classifies every package into a layer and flags each import edge
// whose importer layer does not permit the imported layer. Both endpoints must
// be mapped for an edge to be evaluated: an edge into an unmapped package is
// ignored (the unmapped package surfaces separately as a Warn). Returns sorted,
// deduplicated violation detail lines and the sorted list of unmapped packages.
func checkEdges(pkgs []pkg, m matrix) (violations, unmapped []string) {
	layerOf := map[string]string{}
	unmappedSet := map[string]bool{}
	for _, p := range pkgs {
		l := m.layerFor(p.Path)
		layerOf[p.Path] = l
		if l == "" {
			unmappedSet[p.Path] = true
		}
	}
	vset := map[string]bool{}
	for _, p := range pkgs {
		from := layerOf[p.Path]
		if from == "" {
			continue
		}
		for _, imp := range p.Imports {
			to, ok := layerOf[imp]
			if !ok || to == "" || m.allowed(from, to) {
				continue
			}
			vset[fmt.Sprintf("%s -> %s (%s may not import %s)", p.Path, imp, from, to)] = true
		}
	}
	return sortedKeys(vset), sortedKeys(unmappedSet)
}

func sortedKeys(s map[string]bool) []string {
	out := make([]string, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// stripModulePrefix converts a full import path to a module-relative path,
// returning (rel, true) for in-module packages and ("", false) for stdlib or
// third-party imports. The module prefix is matched on a segment boundary.
func stripModulePrefix(importPath, module string) (string, bool) {
	if importPath == module {
		return "", true
	}
	if strings.HasPrefix(importPath, module+"/") {
		return strings.TrimPrefix(importPath, module+"/"), true
	}
	return "", false
}

// scopePackages converts raw `go list` output into module-scoped pkgs: import
// paths are made module-relative and stdlib/third-party imports are dropped.
// Test imports (in-package and external) fold into the package's import set so
// _test files inherit the package-under-test's layer.
func scopePackages(raw []goListPkg, module string) []pkg {
	var out []pkg
	for _, r := range raw {
		rel, ok := stripModulePrefix(r.ImportPath, module)
		if !ok {
			continue
		}
		seen := map[string]bool{}
		var imps []string
		for _, group := range [][]string{r.Imports, r.TestImports, r.XTestImports} {
			for _, imp := range group {
				irel, ok := stripModulePrefix(imp, module)
				if !ok || irel == rel || seen[irel] {
					continue
				}
				seen[irel] = true
				imps = append(imps, irel)
			}
		}
		out = append(out, pkg{Path: rel, Imports: imps})
	}
	return out
}
