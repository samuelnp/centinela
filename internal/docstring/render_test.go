package docstring

import "testing"

func TestReport_LinesPutsParseErrorsBeforeViolations(t *testing.T) {
	r := Report{
		Violations:  []Violation{{Path: "a.go", Line: 3, Kind: "func", Name: "F"}},
		ParseErrors: []ParseError{{Path: "b.go", Message: "expected declaration"}},
	}
	got := r.Lines()
	if len(got) != 2 {
		t.Fatalf("lines = %v", got)
	}
	if got[0] != "b.go: unparseable Go file: expected declaration" {
		t.Fatalf("first line = %q", got[0])
	}
	if got[1] != "a.go:3: func F has no doc comment" {
		t.Fatalf("second line = %q", got[1])
	}
}

func TestReport_ExemptionLinesNameTheDirective(t *testing.T) {
	r := Report{Exemptions: []Exemption{{Path: "a.go", Line: 9, Kind: "var", Name: "V"}}}
	got := r.ExemptionLines()
	if len(got) != 1 || got[0] != "a.go:9: var V exempt via //centinela:nodoc" {
		t.Fatalf("exemption lines = %v", got)
	}
}

func TestReport_OKIsFalseForEitherFailureKind(t *testing.T) {
	if !(Report{}).OK() {
		t.Fatal("an empty report is OK")
	}
	if (Report{Violations: []Violation{{}}}).OK() {
		t.Fatal("a violation must not be OK")
	}
	if (Report{ParseErrors: []ParseError{{}}}).OK() {
		t.Fatal("a parse error must not be OK")
	}
}

func TestReport_EmptyReportRendersNoLines(t *testing.T) {
	r := Report{}
	if len(r.Lines()) != 0 || len(r.ExemptionLines()) != 0 {
		t.Fatal("an empty report must render no detail lines")
	}
}
