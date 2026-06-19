package importgraph

// pythonProvider builds a Python module import graph by running an embedded AST
// walker via python3, which emits the shared JSON contract (graph_json.go).
// A missing python3 is reported as ToolMissingError (→ gate Warn, CI green).
type pythonProvider struct {
	run Runner
}

func (pythonProvider) Name() string { return "python" }

func (p pythonProvider) Load(root string) (Graph, error) {
	if !onPath("python3") {
		return Graph{}, &ToolMissingError{Tool: "python3"}
	}
	out, err := p.run("python3", "-c", pyASTScript, root)
	if err != nil {
		return Graph{}, err
	}
	return decodeGraphJSON(out)
}
