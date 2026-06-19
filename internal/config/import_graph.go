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
// loads the project's import graph via a language provider and checks every
// intra-project edge against the per-layer allow matrix.
//
// Provider selects the graph backend: "" (auto-detect by manifest), "go",
// "node", "python", or "script". ScriptCommand is the argv run by the "script"
// provider, which must emit the import-graph JSON contract. Module is the Go
// module override (go provider only), defaulting to `go list -m` when blank.
type ImportGraphConfig struct {
	Enabled       bool     `toml:"enabled"`
	Module        string   `toml:"module"`
	Provider      string   `toml:"provider"`
	ScriptCommand []string `toml:"script_command"`
	Layers        []Layer  `toml:"layers"`
}

// NormalizeImportGraph trims whitespace from the module path, provider name,
// script argv, layer names, path globs, and allow entries. It does not drop
// malformed layers: a layer with no paths is preserved so the gate can report
// it as a config error rather than silently ignoring it.
func NormalizeImportGraph(g ImportGraphConfig) ImportGraphConfig {
	g.Module = strings.TrimSpace(g.Module)
	g.Provider = strings.ToLower(strings.TrimSpace(g.Provider))
	g.ScriptCommand = trimNonEmpty(g.ScriptCommand)
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
