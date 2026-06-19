package gates

import (
	"errors"

	"github.com/samuelnp/centinela/internal/importgraph"
)

// classifyLoadError maps a provider load error to a Result. A missing provider
// (no recognized manifest) or a missing external tool is a non-failing Warn —
// the self-skip that stops the non-Go hard-fail and keeps CI green. Any other
// load error (uncompilable code, malformed output, nonzero script exit) is a
// Fail so a real problem is never a silent pass.
func classifyLoadError(err error) Result {
	r := Result{Name: "import_graph", Status: Warn}
	var toolMissing *importgraph.ToolMissingError
	switch {
	case errors.Is(err, importgraph.ErrNoProvider):
		r.Message = "import_graph: no provider matched this project; skipping " +
			"(set gates.import_graph.provider to enforce)."
	case errors.As(err, &toolMissing):
		r.Message = "import_graph: " + toolMissing.Error() + "; skipping."
	default:
		r.Status = Fail
		r.Message = "import_graph: " + err.Error()
	}
	return r
}
