package importgraph

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type depcruiseJSON struct {
	Modules []struct {
		Source       string `json:"source"`
		Dependencies []struct {
			Resolved   string `json:"resolved"`
			CoreModule bool   `json:"coreModule"`
		} `json:"dependencies"`
	} `json:"modules"`
}

// parseDepcruise converts dependency-cruiser JSON into project-scoped Pkgs,
// dropping core modules and anything under node_modules.
func parseDepcruise(out []byte) ([]Pkg, error) {
	var doc depcruiseJSON
	if err := json.Unmarshal(out, &doc); err != nil {
		return nil, fmt.Errorf("decoding dependency-cruiser JSON: %w", err)
	}
	var pkgs []Pkg
	for _, m := range doc.Modules {
		src := normalizeNodePath(m.Source)
		if src == "" {
			continue
		}
		seen := map[string]bool{}
		var imps []string
		for _, d := range m.Dependencies {
			r := normalizeNodePath(d.Resolved)
			if d.CoreModule || r == "" || r == src || seen[r] {
				continue
			}
			seen[r] = true
			imps = append(imps, r)
		}
		pkgs = append(pkgs, Pkg{Path: src, Imports: imps})
	}
	return pkgs, nil
}

// parseMadge converts madge's {file:[deps]} JSON into project-scoped Pkgs,
// sorted by path for deterministic output.
func parseMadge(out []byte) ([]Pkg, error) {
	var doc map[string][]string
	if err := json.Unmarshal(out, &doc); err != nil {
		return nil, fmt.Errorf("decoding madge JSON: %w", err)
	}
	var pkgs []Pkg
	for src, deps := range doc {
		s := normalizeNodePath(src)
		if s == "" {
			continue
		}
		var imps []string
		for _, d := range deps {
			if r := normalizeNodePath(d); r != "" && r != s {
				imps = append(imps, r)
			}
		}
		pkgs = append(pkgs, Pkg{Path: s, Imports: imps})
	}
	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Path < pkgs[j].Path })
	return pkgs, nil
}

func normalizeNodePath(p string) string {
	p = strings.TrimPrefix(strings.ReplaceAll(p, "\\", "/"), "./")
	if p == "" || p == "node_modules" || strings.HasPrefix(p, "node_modules/") ||
		strings.Contains(p, "/node_modules/") {
		return ""
	}
	return p
}
