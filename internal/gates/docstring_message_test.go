package gates

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/docstring"
)

// exemptReport builds a Pass-shaped report with n exemptions.
func exemptReport(n int) docstring.Report {
	rep := docstring.Report{Files: 1, Inspected: n + 1}
	for i := 0; i < n; i++ {
		rep.Exemptions = append(rep.Exemptions, docstring.Exemption{
			Path: "a.go", Line: i + 1, Kind: "func", Name: string(rune('A' + i)),
		})
	}
	return rep
}

// F4 regression: the Pass message must never call an exempted identifier
// documented — RenderGateResult drops Details on Pass, so the message is the
// only place an opt-out is visible on the enforcing surface.
func TestReportDocstring_PassMessageDoesNotClaimExemptedAreDocumented(t *testing.T) {
	r := reportDocstring(exemptReport(1), "fail")
	if r.Status != Pass {
		t.Fatalf("status = %v", r.Status)
	}
	if strings.Contains(r.Message, "All 2 exported identifiers") {
		t.Fatalf("message claims documentation it did not verify: %q", r.Message)
	}
	if !strings.Contains(r.Message, "1 of 2 exported identifiers") ||
		!strings.Contains(r.Message, "1 exempt via //centinela:nodoc") ||
		!strings.Contains(r.Message, "a.go:1 func A") {
		t.Fatalf("message must name the opt-out: %q", r.Message)
	}
}

func TestReportDocstring_PassWithNoExemptionsKeepsTheAllForm(t *testing.T) {
	r := reportDocstring(docstring.Report{Files: 2, Inspected: 7}, "fail")
	if !strings.Contains(r.Message, "All 7 exported identifiers across 2 changed Go file(s) are documented.") {
		t.Fatalf("message = %q", r.Message)
	}
	if len(r.Details) != 0 {
		t.Fatalf("details = %v", r.Details)
	}
}

func TestReportDocstring_ManyExemptionsAreCappedWithACount(t *testing.T) {
	r := reportDocstring(exemptReport(6), "fail")
	if !strings.Contains(r.Message, "+3 more") {
		t.Fatalf("message = %q", r.Message)
	}
	if len(r.Details) != 6 {
		t.Fatalf("every exemption must stay in Details for JSON/audit: %v", r.Details)
	}
}

// F4 regression: a Warn ends with a pointer, never a colon introducing a list
// the renderer will not expand.
func TestReportDocstring_WarnMessageCarriesTheCountNotADanglingColon(t *testing.T) {
	rep := docstring.Report{Files: 1, Inspected: 2,
		Violations: []docstring.Violation{{Path: "a.go", Line: 3, Kind: "func", Name: "F"}}}
	r := reportDocstring(rep, "warn")
	if r.Status != Warn {
		t.Fatalf("status = %v", r.Status)
	}
	if strings.HasSuffix(r.Message, ":") {
		t.Fatalf("warn must not dangle a colon: %q", r.Message)
	}
	if !strings.Contains(r.Message, "1 undocumented exported identifier") ||
		!strings.Contains(r.Message, "centinela docs lint") {
		t.Fatalf("message = %q", r.Message)
	}
}

// Fail keeps the colon: RenderGateResult does expand its Details.
func TestReportDocstring_FailKeepsTheColon(t *testing.T) {
	rep := docstring.Report{Files: 1, Inspected: 1,
		Violations:  []docstring.Violation{{Path: "a.go", Line: 3, Kind: "func", Name: "F"}},
		ParseErrors: []docstring.ParseError{{Path: "b.go", Message: "boom"}}}
	r := reportDocstring(rep, "fail")
	if r.Status != Fail || !strings.HasSuffix(r.Message, ":") {
		t.Fatalf("status=%v message=%q", r.Status, r.Message)
	}
}
