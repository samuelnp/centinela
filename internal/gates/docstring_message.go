package gates

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/docstring"
)

// docstringNamedExemptions caps how many opt-outs the message names inline.
const docstringNamedExemptions = 3

// docstringPassMessage refuses to call an exempted identifier documented, and
// names the opt-outs inline. ui.RenderGateResult expands Result.Details for
// Fail only, so on the enforcing surface (centinela validate, precommit,
// pr-gate) the message is the ONLY place an exemption can be seen — and an
// exemption nobody can see is the hole the Exemption type exists to close.
func docstringPassMessage(rep docstring.Report) string {
	exempt := len(rep.Exemptions)
	if exempt == 0 {
		return fmt.Sprintf(
			"All %d exported identifiers across %d changed Go file(s) are documented.",
			rep.Inspected, rep.Files)
	}
	return fmt.Sprintf(
		"%d of %d exported identifiers across %d changed Go file(s) are documented; %d exempt via %s: %s",
		rep.Inspected-exempt, rep.Inspected, rep.Files, exempt,
		docstring.Directive, summarizeExemptions(rep.Exemptions))
}

// summarizeExemptions names the first few exemptions and counts the rest.
func summarizeExemptions(ex []docstring.Exemption) string {
	names := make([]string, 0, len(ex))
	for i, e := range ex {
		if i == docstringNamedExemptions {
			names = append(names, fmt.Sprintf("+%d more", len(ex)-i))
			break
		}
		names = append(names, fmt.Sprintf("%s:%d %s %s", e.Path, e.Line, e.Kind, e.Name))
	}
	return strings.Join(names, ", ")
}

// docstringProblemMessage carries the counts in the message itself. Only Fail
// gets its Details expanded by the renderer, so a Warn must not end in a colon
// introducing a list nobody will ever see — it points at the surface that can.
//
// Violations and parse errors are counted SEPARATELY. Summing them under one
// "undocumented identifiers" label made a scope whose only problem was an
// unparseable file report "1 undocumented" when the true count was zero — and
// on Warn that false sentence is the operator's entire output.
func docstringProblemMessage(rep docstring.Report, status Status) string {
	var parts []string
	if n := len(rep.Violations); n > 0 {
		parts = append(parts, fmt.Sprintf("%d undocumented exported identifier%s", n, plural(n)))
	}
	if n := len(rep.ParseErrors); n > 0 {
		parts = append(parts, fmt.Sprintf("%d unparseable file%s", n, plural(n)))
	}
	body := strings.Join(parts, ", ") + " in changed files"
	if status == Fail {
		return body + ":"
	}
	return body + " — run `centinela docs lint` to list them."
}

// plural returns the "s" suffix for a count.
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
