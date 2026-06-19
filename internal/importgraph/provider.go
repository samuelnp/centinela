// Package importgraph is a leaf that loads a project's intra-project import
// graph for the G2 gate across languages. It defines the GraphProvider seam,
// the per-language backends (go/node/python) and a custom-script escape hatch,
// and manifest-based provider selection. Every backend returns an already
// project-scoped Graph so the gate's matrix logic is language-agnostic.
//
// It depends only on the standard library, os/exec, and the internal/golist
// leaf (used by the Go backend) — keeping it a leaf importable by the gate.
package importgraph

import "errors"

// Pkg is a single logical unit (package/module/directory) and its intra-project
// import targets, already made project-relative (e.g. "internal/config").
type Pkg struct {
	Path    string
	Imports []string
}

// Graph is a project's scoped import graph: a root identifier plus its packages.
type Graph struct {
	Module string
	Pkgs   []Pkg
}

// GraphProvider loads the scoped import graph for one language/ecosystem.
type GraphProvider interface {
	Name() string
	Load(root string) (Graph, error)
}

// Runner executes an external command and returns stdout, folding the first
// stderr line into the error. Injected so backends are unit-testable without
// the real toolchain installed.
type Runner func(name string, args ...string) ([]byte, error)

// ErrNoProvider is returned by Select when no language matched the project; the
// gate maps it to a non-failing self-skip Warn (the non-Go hard-fail fix).
var ErrNoProvider = errors.New("no import-graph provider matched this project")

// ToolMissingError reports that a backend's required external tool is absent on
// PATH. The gate maps it to a non-failing Warn so CI without the tool stays
// green; install the tool or switch to provider="script" to enforce.
type ToolMissingError struct{ Tool string }

func (e *ToolMissingError) Error() string {
	return e.Tool + " is not installed (install it or set gates.import_graph.provider)"
}
