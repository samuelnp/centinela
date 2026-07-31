package gates

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/docstring"
)

// reportDocstring maps a scan report onto a gate Result. A scan that opened no
// file after generated-file filtering is a Skip, not a green pass: every
// in-scope file may have been generated, and a gate that inspected nothing may
// not claim success.
func reportDocstring(rep docstring.Report, severity string) Result {
	r := Result{Name: docstringGate}
	if rep.Files == 0 && rep.OK() {
		r.Status = Skip
		r.Message = "No Go files in scope — nothing inspected."
		return r
	}
	if rep.OK() {
		r.Status = Pass
		r.Message = fmt.Sprintf(
			"All %d exported identifiers across %d changed Go file(s) are documented.",
			rep.Inspected, rep.Files)
		r.Details = rep.ExemptionLines()
		return r
	}
	if severity == "warn" {
		r.Status = Warn
	} else {
		r.Status = Fail
	}
	r.Message = "Undocumented exported identifiers in changed files:"
	r.Details = append(rep.Lines(), rep.ExemptionLines()...)
	return r
}
