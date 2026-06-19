package importgraph

import "fmt"

// Select resolves the GraphProvider for a project. An explicit provider kind
// (go|node|python|script) wins; otherwise the kind is auto-detected from the
// project's manifest. Returns ErrNoProvider when nothing matches (→ gate
// self-skips with a Warn). "script" is never auto-selected and requires a
// non-empty scriptCmd. module is the Go module override (ignored by others);
// run is the injected command Runner (nil → the default exec runner).
func Select(root, provider, module string, scriptCmd []string, run Runner) (GraphProvider, error) {
	if run == nil {
		run = execRunner
	}
	kind := provider
	if kind == "" {
		kind = detectKind(root)
	}
	switch kind {
	case "go":
		return goProvider{module: module}, nil
	case "node":
		return nodeProvider{run: run}, nil
	case "python":
		return pythonProvider{run: run}, nil
	case "script":
		if len(scriptCmd) == 0 {
			return nil, fmt.Errorf("provider \"script\" requires gates.import_graph.script_command")
		}
		return scriptProvider{cmd: scriptCmd, run: run}, nil
	case "":
		return nil, ErrNoProvider
	default:
		return nil, fmt.Errorf("unknown import-graph provider %q", kind)
	}
}
