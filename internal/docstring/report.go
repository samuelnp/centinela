package docstring

// Violation is one exported identifier declared without a doc comment.
type Violation struct {
	// Path is the slash-normalized file the identifier is declared in.
	Path string
	// Line is the 1-based declaration line.
	Line int
	// Kind is the declaration kind: func, method, type, const, var,
	// interface method or field.
	Kind string
	// Name is the exported identifier.
	Name string
}

// Exemption is an identifier excused from the check by a Directive comment.
// Exemptions are reported even on a passing run: an opt-out nobody can see is
// a hole.
type Exemption struct {
	// Path is the slash-normalized file the identifier is declared in.
	Path string
	// Line is the 1-based declaration line.
	Line int
	// Kind is the declaration kind.
	Kind string
	// Name is the exempted identifier.
	Name string
}

// ParseError records a file the scanner could not parse. It is reported as a
// violation of its own kind and never as a silent skip — a file that vanished
// from the scan is exactly the false green truthful-validators forbids.
type ParseError struct {
	// Path is the slash-normalized file that failed to parse.
	Path string
	// Message is the first line of the parser error.
	Message string
}

// Report is the outcome of one scan.
type Report struct {
	// Files is the number of files actually parsed and walked.
	Files int
	// Inspected is the number of exported identifiers examined.
	Inspected int
	// Violations are the undocumented exported identifiers found.
	Violations []Violation
	// Exemptions are the identifiers excused by Directive.
	Exemptions []Exemption
	// ParseErrors are the files that could not be parsed.
	ParseErrors []ParseError
}

// OK reports whether the scan found neither a violation nor a parse error.
func (r Report) OK() bool {
	return len(r.Violations) == 0 && len(r.ParseErrors) == 0
}
