package analyze

import "sort"

// hasManifest reports whether any detected manifest is of the given kind.
func hasManifest(manifests []Manifest, kind string) bool {
	for _, m := range manifests {
		if m.Kind == kind {
			return true
		}
	}
	return false
}

// declaredEdges builds edges from each manifest to its declared dependency
// names. The "from" node is the manifest path so callers can distinguish which
// ecosystem declared a dependency. Edges are de-duplicated and sorted.
func declaredEdges(manifests []Manifest) []Edge {
	seen := map[Edge]bool{}
	for _, m := range manifests {
		for _, dep := range m.Deps {
			seen[Edge{From: m.Path, To: dep}] = true
		}
	}
	return sortedEdges(seen)
}

// sortedEdges returns the edge set sorted by From then To. It always returns a
// non-nil slice so the JSON edges field is "[]" not "null".
func sortedEdges(set map[Edge]bool) []Edge {
	out := make([]Edge, 0, len(set))
	for e := range set {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].To < out[j].To
	})
	return out
}
