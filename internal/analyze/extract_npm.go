package analyze

import (
	"encoding/json"
	"sort"
)

// packageJSON is the subset of package.json the npm extractor reads. Decoding
// into concrete types (no map[string]any in the output path) keeps the result
// typed and deterministic.
type packageJSON struct {
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// extractPackageJSON records the build/test scripts and declared dependency
// names. Invalid JSON leaves m as detected-but-unparsable (no signals), so the
// overall scan continues (AC-4).
func extractPackageJSON(data []byte, m *Manifest) {
	var pj packageJSON
	if err := json.Unmarshal(data, &pj); err != nil {
		return
	}
	m.Build = pj.Scripts["build"]
	m.Test = pj.Scripts["test"]
	m.Deps = sortedDepNames(pj.Dependencies, pj.DevDependencies)
	m.Framework = npmFramework(pj.Dependencies, pj.DevDependencies)
}

// sortedDepNames returns the unique, sorted dependency names across the given
// dependency maps.
func sortedDepNames(maps ...map[string]string) []string {
	seen := map[string]bool{}
	for _, mp := range maps {
		for name := range mp {
			seen[name] = true
		}
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	if len(out) == 0 {
		return nil
	}
	return out
}

// npmFramework names a well-known framework when its package is a declared
// dependency, checked in priority order so the result is deterministic.
func npmFramework(maps ...map[string]string) string {
	has := func(pkg string) bool {
		for _, mp := range maps {
			if _, ok := mp[pkg]; ok {
				return true
			}
		}
		return false
	}
	for _, fw := range []struct{ pkg, name string }{
		{"next", "Next.js"}, {"react", "React"},
		{"vue", "Vue"}, {"@angular/core", "Angular"},
		{"svelte", "Svelte"}, {"express", "Express"},
	} {
		if has(fw.pkg) {
			return fw.name
		}
	}
	return ""
}
