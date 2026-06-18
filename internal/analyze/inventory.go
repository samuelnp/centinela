// Package analyze produces a deterministic, read-only inventory of a codebase
// (languages, manifests, locales, package layout, dependency graph) with no LLM
// call. It is the substrate downstream Phase 9 features bind to via the
// schemaVersion-tagged Inventory contract written to .workflow/analysis.json.
package analyze

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SchemaVersion is the contract version of the Inventory JSON. Any change to the
// serialized field set after v1 bumps this so downstream consumers can detect
// format drift.
const SchemaVersion = 1

// DefaultOutPath is the well-known committed location of the inventory. It lives
// here (not in cmd/) so downstream features import the same constant.
const DefaultOutPath = ".workflow/analysis.json"

// Inventory is the complete machine-readable description of a repository. Every
// slice is pre-sorted before Save so serialization is byte-stable.
type Inventory struct {
	SchemaVersion   int             `json:"schemaVersion"`
	PrimaryLanguage string          `json:"primaryLanguage"`
	Languages       []LanguageStat  `json:"languages"`
	Manifests       []Manifest      `json:"manifests"`
	Locales         []string        `json:"locales"`
	Packages        []string        `json:"packages"`
	Graph           DependencyGraph `json:"graph"`
}

// LanguageStat is one detected language and how many source files matched it.
type LanguageStat struct {
	Name      string `json:"name"`
	FileCount int    `json:"fileCount"`
}

// Manifest is one detected build/dependency manifest and its extracted signals.
type Manifest struct {
	Kind      string   `json:"kind"`
	Path      string   `json:"path"`
	Framework string   `json:"framework,omitempty"`
	Build     string   `json:"build,omitempty"`
	Test      string   `json:"test,omitempty"`
	Deps      []string `json:"deps,omitempty"`
}

// DependencyGraph is the package/declared-dep edge set for the repo.
type DependencyGraph struct {
	Kind   string `json:"kind"`
	Module string `json:"module,omitempty"`
	Edges  []Edge `json:"edges"`
	Note   string `json:"note,omitempty"`
}

// Edge is a single directed dependency edge.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Save writes the inventory to path as indented JSON with a trailing newline.
// The Inventory must already have sorted slices; Save adds no ordering of its
// own so output is byte-identical across re-runs of the same value. A write
// failure (e.g. un-writable path) returns an error and leaves no partial file:
// the full payload is marshaled in memory first.
func Save(path string, inv Inventory) error {
	data, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}
