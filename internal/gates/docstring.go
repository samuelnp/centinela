package gates

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/docstring"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// docstringGate is the gate's reported name.
const docstringGate = "docstring-gate"

// CheckDocstring runs the docstring gate standalone and also returns the
// underlying scan report, so `centinela docs lint` can print counts without a
// second implementation of the check.
func CheckDocstring(cfg *config.Config, filter *gitdiff.Set) (Result, docstring.Report) {
	return docstringCheck(cfg, filter)
}

// checkDocstring reports exported identifiers in changed files that carry no
// doc comment. It is a ratchet: the scope is always the changed-file set, so a
// legacy backlog in untouched files is never opened and severity "fail" cannot
// turn main red for a pre-existing gap. Touch a file, document its exports.
func checkDocstring(cfg *config.Config, filter *gitdiff.Set) Result {
	r, _ := docstringCheck(cfg, filter)
	return r
}

func docstringCheck(cfg *config.Config, filter *gitdiff.Set) (Result, docstring.Report) {
	files, skip := docstringFiles(cfg, filter)
	if skip != "" {
		return Result{Name: docstringGate, Status: Skip, Message: skip}, docstring.Report{}
	}
	scanner, ok := docstring.For(docstring.GoLang)
	if !ok {
		return Result{
			Name:    docstringGate,
			Status:  Skip,
			Message: "No docstring scanner registered for Go — nothing inspected.",
		}, docstring.Report{}
	}
	rep, err := scanner.Scan(files, DocstringOptions(cfg))
	if err != nil {
		return Result{
			Name:    docstringGate,
			Status:  Fail,
			Message: "docstring: " + err.Error(),
		}, rep
	}
	return reportDocstring(rep, cfg.Gates.Docstring.Severity), rep
}

// DocstringOptions maps the [gates.docstring] config block onto the scanner's
// Options. Exported so `centinela docs lint --full` builds its report scope
// from the same knobs the gate enforces with, never a second interpretation.
func DocstringOptions(cfg *config.Config) docstring.Options {
	c := cfg.Gates.Docstring
	return docstring.Options{
		Roots:             c.Roots,
		IncludeInternal:   c.IncludesInternal(),
		CheckFields:       c.CheckFields,
		RequireNamePrefix: c.RequireNamePrefix,
	}
}
