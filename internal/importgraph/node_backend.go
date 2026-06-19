package importgraph

// nodeProvider builds a JS/TS import graph by shelling out to dependency-cruiser
// (preferred) or madge and parsing their JSON. When neither tool is on PATH it
// returns ToolMissingError (→ gate Warn, CI green). Paths are project-relative
// file paths matched against the gate's layer globs.
type nodeProvider struct {
	run Runner
}

func (nodeProvider) Name() string { return "node" }

func (p nodeProvider) Load(root string) (Graph, error) {
	switch {
	case onPath("depcruise"):
		out, err := p.run("depcruise", "--output-type", "json", "--no-config", root)
		if err != nil {
			return Graph{}, err
		}
		pkgs, err := parseDepcruise(out)
		return Graph{Module: "node", Pkgs: pkgs}, err
	case onPath("madge"):
		out, err := p.run("madge", "--json", root)
		if err != nil {
			return Graph{}, err
		}
		pkgs, err := parseMadge(out)
		return Graph{Module: "node", Pkgs: pkgs}, err
	default:
		return Graph{}, &ToolMissingError{Tool: "dependency-cruiser or madge"}
	}
}
