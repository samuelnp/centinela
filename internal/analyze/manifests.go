package analyze

import (
	"os"
	"path/filepath"
	"sort"
)

// extractor parses a manifest's bytes into the signal fields of m. A malformed
// manifest returns m unchanged (detected-but-unparsable) — it never errors, so
// the scan continues (AC-4).
type extractor func(data []byte, m *Manifest)

// manifestEntry binds a manifest filename to its Kind and signal extractor.
type manifestEntry struct {
	kind    string
	extract extractor
}

// manifestTable maps a manifest filename to its detection entry. Adding an
// ecosystem is a table edit.
var manifestTable = map[string]manifestEntry{
	"go.mod":           {kind: "go-mod", extract: extractGoMod},
	"package.json":     {kind: "npm", extract: extractPackageJSON},
	"Cargo.toml":       {kind: "cargo", extract: extractCargo},
	"Gemfile":          {kind: "gem", extract: extractGemfile},
	"pyproject.toml":   {kind: "python", extract: extractPyproject},
	"requirements.txt": {kind: "python", extract: extractRequirements},
	"Makefile":         {kind: "make", extract: extractMakefile},
}

// detectManifests scans the root directory (top level only) for known manifest
// files, runs each extractor best-effort, and returns the manifests sorted by
// path. Detection never fails: an unreadable or malformed file still yields a
// detected manifest with empty signals.
func detectManifests(root string) []Manifest {
	var out []Manifest
	for name, entry := range manifestTable {
		path := filepath.Join(root, name)
		if _, err := os.Stat(path); err != nil {
			continue
		}
		m := Manifest{Kind: entry.kind, Path: name}
		if data, err := os.ReadFile(path); err == nil {
			entry.extract(data, &m)
			sort.Strings(m.Deps)
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}
