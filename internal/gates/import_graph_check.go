package gates

import (
	"fmt"
	"sort"
)

// pkg is a single package and its intra-project imports, already scoped to the
// project (paths are relative to the project root, e.g. "internal/config").
// Adapted from the provider's graph by the loader; consumed by pure check logic.
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
