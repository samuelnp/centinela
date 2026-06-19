package importgraph

import (
	"strings"

	"github.com/samuelnp/centinela/internal/golist"
)

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

// scopeGoPkgs converts raw `go list` output into module-scoped Pkgs: import
// paths are made module-relative and stdlib/third-party imports are dropped.
// Test imports (in-package and external) fold into the import set so _test files
// inherit the package-under-test's layer and cannot hide a forbidden edge.
func scopeGoPkgs(raw []golist.Pkg, module string) []Pkg {
	var out []Pkg
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
		out = append(out, Pkg{Path: rel, Imports: imps})
	}
	return out
}
