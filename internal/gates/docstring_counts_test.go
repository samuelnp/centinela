package gates

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/docstring"
)

// Finding 2 regression: an unparseable file is not an undocumented identifier.
// Summing the two under one label made a scope whose only problem was a broken
// file report "1 undocumented" when the true count was zero — and on Warn that
// sentence is the operator's entire output.
func TestReportDocstring_ParseErrorsAreNotCountedAsUndocumented(t *testing.T) {
	rep := docstring.Report{Files: 1, Inspected: 3,
		ParseErrors: []docstring.ParseError{{Path: "b.go", Message: "boom"}}}
	r := reportDocstring(rep, "warn")
	if r.Status != Warn {
		t.Fatalf("status = %v", r.Status)
	}
	if strings.Contains(r.Message, "undocumented") {
		t.Fatalf("zero violations must not be reported as undocumented: %q", r.Message)
	}
	if !strings.Contains(r.Message, "1 unparseable file in changed files") {
		t.Fatalf("message = %q", r.Message)
	}
}

func TestReportDocstring_BothCategoriesAreNamedSeparately(t *testing.T) {
	rep := docstring.Report{Files: 2, Inspected: 4,
		Violations: []docstring.Violation{
			{Path: "a.go", Line: 3, Kind: "func", Name: "F"},
			{Path: "a.go", Line: 9, Kind: "var", Name: "V"}},
		ParseErrors: []docstring.ParseError{{Path: "b.go", Message: "boom"}}}
	r := reportDocstring(rep, "fail")
	want := "2 undocumented exported identifiers, 1 unparseable file in changed files:"
	if r.Message != want {
		t.Fatalf("message = %q, want %q", r.Message, want)
	}
}

func TestReportDocstring_SingularAndPluralAgreeWithTheCount(t *testing.T) {
	one := reportDocstring(docstring.Report{Files: 1, Inspected: 1,
		Violations: []docstring.Violation{{Path: "a.go", Line: 3, Kind: "func", Name: "F"}}}, "fail")
	if !strings.Contains(one.Message, "1 undocumented exported identifier in") {
		t.Fatalf("singular: %q", one.Message)
	}
	two := reportDocstring(docstring.Report{Files: 1, Inspected: 2,
		Violations: []docstring.Violation{
			{Path: "a.go", Line: 3}, {Path: "a.go", Line: 4}}}, "fail")
	if !strings.Contains(two.Message, "2 undocumented exported identifiers") {
		t.Fatalf("plural: %q", two.Message)
	}
}
