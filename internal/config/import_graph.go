package config

import "strings"

// Layer is one entry in the import-graph allow matrix: a named group of
// packages (matched by path globs relative to the module root) plus the list
// of other layer names its packages are permitted to import.
type Layer struct {
	Name  string   `toml:"name"`
	Paths []string `toml:"paths"`
	Allow []string `toml:"allow"`
}

// ImportGraphConfig controls the G2 import-graph gate. When Enabled, the gate
// loads the whole module's import graph and checks every intra-module edge
// against the per-layer allow matrix. Module defaults to the path reported by
// `go list -m` when left blank.
type ImportGraphConfig struct {
	Enabled bool    `toml:"enabled"`
	Module  string  `toml:"module"`
	Layers  []Layer `toml:"layers"`
}

// NormalizeImportGraph trims whitespace from the module path, layer names,
// path globs, and allow entries. It does not drop malformed layers: a layer
// with no paths is preserved so the gate can report it as a config error
// rather than silently ignoring it.
func NormalizeImportGraph(g ImportGraphConfig) ImportGraphConfig {
	g.Module = strings.TrimSpace(g.Module)
	cleaned := make([]Layer, 0, len(g.Layers))
	for _, l := range g.Layers {
		l.Name = strings.TrimSpace(l.Name)
		l.Paths = trimNonEmpty(l.Paths)
		l.Allow = trimNonEmpty(l.Allow)
		cleaned = append(cleaned, l)
	}
	g.Layers = cleaned
	return g
}

func trimNonEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}
