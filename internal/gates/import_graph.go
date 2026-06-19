package gates

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// checkImportGraph enforces the G2 layer-boundary rule by loading the project's
// import graph via a language provider (Go/Node/Python/custom-script, selected
// by manifest unless configured) and checking every intra-project edge against
// the configured allow matrix.
//
// The filter argument is accepted for RunWithFilter signature parity but is
// DELIBERATELY IGNORED: a forbidden edge can be introduced or removed by a file
// outside the current diff set, so a diff-scoped load would produce false
// passes. This gate always loads the entire project.
func checkImportGraph(cfg *config.Config, _ *gitdiff.Set) Result {
	g := cfg.Gates.ImportGraph
	if len(g.Layers) == 0 {
		return Result{Name: "import_graph", Status: Warn, Message: "import_graph: layer matrix is empty."}
	}
	m, err := buildMatrix(g.Layers)
	if err != nil {
		return Result{Name: "import_graph", Status: Fail, Message: "import_graph config: " + err.Error()}
	}
	graph, err := loadGraph(g)
	if err != nil {
		return classifyLoadError(err)
	}
	violations, unmapped := checkEdges(toPkgs(graph.Pkgs), m)
	return reportEdges(violations, unmapped)
}

// reportEdges maps the edge-check outcome to a Result: forbidden edges -> Fail
// (one detail per edge); only unmapped packages -> Warn; otherwise Pass.
func reportEdges(violations, unmapped []string) Result {
	r := Result{Name: "import_graph"}
	switch {
	case len(violations) > 0:
		r.Status = Fail
		r.Message = "Forbidden cross-layer imports detected:"
		r.Details = violations
	case len(unmapped) > 0:
		r.Status = Warn
		r.Message = "Packages match no configured layer:"
		r.Details = unmapped
	default:
		r.Status = Pass
		r.Message = "All intra-project imports respect the layer matrix."
	}
	return r
}
