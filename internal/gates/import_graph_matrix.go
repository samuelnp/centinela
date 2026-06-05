package gates

import (
	"fmt"
	"path"

	"github.com/samuelnp/centinela/internal/config"
)

// matrix is the compiled, validated layer allow-matrix: an ordered list of
// layers (each with its path globs) and a per-layer set of layer names it may
// import. It is pure data built by buildMatrix from config; no I/O.
type matrix struct {
	layers []config.Layer
	allow  map[string]map[string]bool
}

// buildMatrix validates the configured layers and compiles the allow-sets.
// Returns an error (surfaced by the gate as a config error) when a layer has
// no paths or an allow entry references an unknown layer name. Duplicate layer
// names are permitted: their paths and allow entries are unioned, mirroring the
// spec where "domain" appears twice (workflow + gates).
func buildMatrix(layers []config.Layer) (matrix, error) {
	names := map[string]bool{}
	for _, l := range layers {
		if l.Name == "" {
			return matrix{}, fmt.Errorf("a layer has an empty name")
		}
		if len(l.Paths) == 0 {
			return matrix{}, fmt.Errorf("layer %q has no paths", l.Name)
		}
		names[l.Name] = true
	}
	allow := map[string]map[string]bool{}
	for _, l := range layers {
		if allow[l.Name] == nil {
			allow[l.Name] = map[string]bool{}
		}
		for _, a := range l.Allow {
			if !names[a] {
				return matrix{}, fmt.Errorf("layer %q allows unknown layer %q", l.Name, a)
			}
			allow[l.Name][a] = true
		}
	}
	return matrix{layers: layers, allow: allow}, nil
}

// layerFor returns the layer name a package belongs to by matching its path
// (relative to the module root) against each layer's globs, or "" when no
// layer matches. First matching layer in config order wins.
func (m matrix) layerFor(rel string) string {
	for _, l := range m.layers {
		for _, g := range l.Paths {
			if globMatch(g, rel) {
				return l.Name
			}
		}
	}
	return ""
}

// globMatch matches a "/"-separated path against a glob where a trailing "/**"
// matches the directory and everything beneath it; otherwise path.Match
// semantics apply per segment.
func globMatch(glob, rel string) bool {
	if base, ok := trimDoubleStar(glob); ok {
		return rel == base || hasPrefixDir(rel, base)
	}
	ok, err := path.Match(glob, rel)
	return err == nil && ok
}
