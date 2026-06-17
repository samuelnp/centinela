package gates

import (
	"strings"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// checkImportGraph enforces the G2 layer-boundary rule by loading the whole
// module's import graph and checking every intra-module edge against the
// configured allow matrix.
//
// The filter argument is accepted for RunWithFilter signature parity but is
// DELIBERATELY IGNORED: a forbidden edge can be introduced or removed by a file
// outside the current diff set, so a diff-scoped load would produce false
// passes. This gate always loads the entire module.
func checkImportGraph(cfg *config.Config, _ *gitdiff.Set) Result {
	r := Result{Name: "import_graph"}
	g := cfg.Gates.ImportGraph
	if len(g.Layers) == 0 {
		r.Status = Warn
		r.Message = "import_graph: layer matrix is empty."
		return r
	}
	m, err := buildMatrix(g.Layers)
	if err != nil {
		r.Status = Fail
		r.Message = "import_graph config: " + err.Error()
		return r
	}
	module, err := resolveModule(g.Module)
	if err != nil {
		r.Status = Fail
		r.Message = "import_graph config: " + err.Error()
		return r
	}
	return runImportGraph(m, module)
}

// resolveModule returns the configured module path, or discovers it via
// `go list -m` when blank. A configured-but-empty module is a config error.
func resolveModule(configured string) (string, error) {
	if strings.TrimSpace(configured) != "" {
		return configured, nil
	}
	module, err := loadModulePath()
	if err != nil {
		return "", err
	}
	if module == "" {
		return "", errEmptyModule
	}
	return module, nil
}

// runImportGraph performs the load -> scope -> check pipeline and maps the
// outcome to a Result: load error -> Fail; forbidden edges -> Fail (one detail
// per edge); only unmapped packages -> Warn; otherwise Pass.
func runImportGraph(m matrix, module string) Result {
	r := Result{Name: "import_graph"}
	raw, err := loadPackages()
	if err != nil {
		r.Status = Fail
		r.Message = "import_graph: " + err.Error()
		return r
	}
	violations, unmapped := checkEdges(scopePackages(raw, module), m)
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
		r.Message = "All intra-module imports respect the layer matrix."
	}
	return r
}
