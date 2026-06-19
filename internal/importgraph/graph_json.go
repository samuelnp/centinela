package importgraph

import (
	"encoding/json"
	"fmt"
)

// graphJSON is the wire contract emitted by the python backend's AST walker and
// by any custom-script provider:
//
//	{"module":"<name>","pkgs":[{"path":"a","imports":["b"]}]}
//
// Paths and imports must already be project-relative logical units that match
// the gate's layer globs.
type graphJSON struct {
	Module string `json:"module"`
	Pkgs   []struct {
		Path    string   `json:"path"`
		Imports []string `json:"imports"`
	} `json:"pkgs"`
}

// decodeGraphJSON parses the shared contract into a Graph. Malformed JSON is an
// error (never a silent empty graph); a valid document with no pkgs is a valid
// empty graph (→ Pass).
func decodeGraphJSON(out []byte) (Graph, error) {
	var doc graphJSON
	if err := json.Unmarshal(out, &doc); err != nil {
		return Graph{}, fmt.Errorf("decoding import-graph JSON: %w", err)
	}
	g := Graph{Module: doc.Module}
	for _, p := range doc.Pkgs {
		g.Pkgs = append(g.Pkgs, Pkg{Path: p.Path, Imports: p.Imports})
	}
	return g, nil
}
