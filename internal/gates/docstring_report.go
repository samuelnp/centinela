package gates

import "github.com/samuelnp/centinela/internal/docstring"

// reportDocstring maps a scan report onto a gate Result. A scan that opened no
// file after generated-file filtering is a Skip, not a green pass: every
// in-scope file may have been generated, and a gate that inspected nothing may
// not claim success.
func reportDocstring(rep docstring.Report, severity string) Result {
	r := Result{Name: docstringGate}
	if rep.Files == 0 && rep.OK() {
		// Distinct from the roots-based skip in docstringFiles: files WERE in
		// scope, the scanner just opened none of them. Saying "no Go files in
		// scope" here would report the wrong cause.
		r.Status = Skip
		r.Message = "No Go files inspected — every file in scope was generated " +
			"or no longer on disk."
		return r
	}
	if rep.OK() {
		r.Status = Pass
		r.Message = docstringPassMessage(rep)
		r.Details = rep.ExemptionLines()
		return r
	}
	if severity == "warn" {
		r.Status = Warn
	} else {
		r.Status = Fail
	}
	r.Message = docstringProblemMessage(rep, r.Status)
	r.Details = append(rep.Lines(), rep.ExemptionLines()...)
	return r
}
