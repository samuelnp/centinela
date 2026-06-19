package importgraph

// scriptProvider is the escape hatch for languages without a built-in backend:
// the user configures a command (argv) that, run at the project root, emits the
// shared import-graph JSON contract on stdout (see graph_json.go). A non-zero
// exit is surfaced as an error (→ gate Fail); valid empty output → empty graph.
type scriptProvider struct {
	cmd []string
	run Runner
}

func (scriptProvider) Name() string { return "script" }

func (p scriptProvider) Load(string) (Graph, error) {
	out, err := p.run(p.cmd[0], p.cmd[1:]...)
	if err != nil {
		return Graph{}, err
	}
	return decodeGraphJSON(out)
}
