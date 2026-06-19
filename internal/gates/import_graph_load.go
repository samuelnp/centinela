package gates

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/importgraph"
)

// loadGraph resolves the import-graph provider for the project (manifest-driven
// unless cfg.Provider is set) and loads the scoped graph. The project root is
// the current working directory — the gate is invoked from the project root.
func loadGraph(cfg config.ImportGraphConfig) (importgraph.Graph, error) {
	provider, err := importgraph.Select(".", cfg.Provider, cfg.Module, cfg.ScriptCommand, nil)
	if err != nil {
		return importgraph.Graph{}, err
	}
	return provider.Load(".")
}

// toPkgs adapts importgraph.Pkg values to the gate-local pkg type so checkEdges
// and its existing tests stay unchanged regardless of the source provider.
func toPkgs(in []importgraph.Pkg) []pkg {
	out := make([]pkg, 0, len(in))
	for _, p := range in {
		out = append(out, pkg{Path: p.Path, Imports: p.Imports})
	}
	return out
}
