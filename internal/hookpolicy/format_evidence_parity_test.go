package hookpolicy_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/hookpolicy"
)

// TestEvidenceKeyOrderParity is the behaviour-level guard the
// format_evidence_order.go doc comment promises: it byte-compares the
// hookpolicy formatter's key ordering against canonical
// evidence.MarshalJSON output for an evidence doc that carries a coverage
// field. If the hookpolicy jsonKeyOrder ever drifts from the evidence one
// (e.g. drops "coverage"), the byte comparison fails.
func TestEvidenceKeyOrderParity(t *testing.T) {
	cov := 87.5
	mobile := true
	re := &evidence.RoleEvidence{
		Feature:     "code-quality-hardening",
		Step:        "tests",
		Role:        "qa-senior",
		Status:      "done",
		GeneratedAt: "2026-06-10T00:00:00Z",
		Inputs:      []string{"docs/plans/x.md"},
		Outputs:     []string{"tests/acceptance/x_test.go"},
		EdgeCases:   []string{"gofmt exits 0"},
		MobileFirst: &mobile,
		Coverage:    &cov,
		HandoffTo:   "validation-specialist",
	}
	canonical, err := re.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	// Feed canonical bytes through the postwrite formatter. Because they are
	// already canonical, FormatEvidence reports no change and returns them
	// byte-for-byte — proving the two key orderings agree.
	path := ".workflow/code-quality-hardening-qa-senior.json"
	out, changed, ferr := hookpolicy.FormatEvidence(path, canonical, "code-quality-hardening")
	if ferr != nil {
		t.Fatalf("FormatEvidence: %v", ferr)
	}
	if changed {
		t.Fatalf("formatter reordered canonical output:\ncanonical:\n%s\nformatter:\n%s", canonical, out)
	}
	if !bytes.Equal(out, canonical) {
		t.Fatalf("formatter output not byte-identical to canonical:\n%s\nvs\n%s", out, canonical)
	}

	// Coverage must land between mobileFirst and handoffTo.
	s := string(canonical)
	mi := strings.Index(s, `"mobileFirst"`)
	ci := strings.Index(s, `"coverage"`)
	hi := strings.Index(s, `"handoffTo"`)
	if mi < 0 || ci < 0 || hi < 0 || !(mi < ci && ci < hi) {
		t.Fatalf("coverage key not between mobileFirst and handoffTo: mi=%d ci=%d hi=%d\n%s", mi, ci, hi, s)
	}
}
