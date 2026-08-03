package docstring

import "fmt"

// String renders a violation as one gate detail line.
func (v Violation) String() string {
	return fmt.Sprintf("%s:%d: %s %s has no doc comment", v.Path, v.Line, v.Kind, v.Name)
}

// String renders an exemption as one gate detail line.
func (e Exemption) String() string {
	return fmt.Sprintf("%s:%d: %s %s exempt via %s", e.Path, e.Line, e.Kind, e.Name, Directive)
}

// String renders a parse error as one gate detail line.
func (p ParseError) String() string {
	return fmt.Sprintf("%s: unparseable Go file: %s", p.Path, p.Message)
}

// Lines renders every parse error and violation as gate detail lines, parse
// errors first because an unreadable file invalidates the rest of the report.
func (r Report) Lines() []string {
	out := make([]string, 0, len(r.ParseErrors)+len(r.Violations))
	for _, p := range r.ParseErrors {
		out = append(out, p.String())
	}
	for _, v := range r.Violations {
		out = append(out, v.String())
	}
	return out
}

// ExemptionLines renders every exemption as a gate detail line.
func (r Report) ExemptionLines() []string {
	out := make([]string, 0, len(r.Exemptions))
	for _, e := range r.Exemptions {
		out = append(out, e.String())
	}
	return out
}
