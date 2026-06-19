package importgraph

import (
	"errors"
	"strings"

	"github.com/samuelnp/centinela/internal/golist"
)

// goProvider is the reference backend: it loads the module's package graph via
// the shared internal/golist seam and scopes it to module-relative Pkgs. module
// is the configured override; when blank it is discovered via `go list -m`.
type goProvider struct {
	module string
}

func (goProvider) Name() string { return "go" }

func (p goProvider) Load(string) (Graph, error) {
	module, err := p.resolveModule()
	if err != nil {
		return Graph{}, err
	}
	raw, err := golist.Packages()
	if err != nil {
		return Graph{}, err
	}
	return Graph{Module: module, Pkgs: scopeGoPkgs(raw, module)}, nil
}

// resolveModule returns the configured module path, or discovers it via
// `go list -m` when blank. A discovered-but-empty module is an error.
func (p goProvider) resolveModule() (string, error) {
	if strings.TrimSpace(p.module) != "" {
		return p.module, nil
	}
	module, err := golist.ModulePath()
	if err != nil {
		return "", err
	}
	if module == "" {
		return "", errors.New("module path is empty")
	}
	return module, nil
}
