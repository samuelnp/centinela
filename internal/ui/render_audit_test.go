package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/audit"
)

// TestRenderAuditDiff_NewAndResolved covers the header counts plus both
// auditSection branches (New blocking set + Resolved prunable set). With new
// violations present, HasNew is true so the "no new" line is suppressed.
func TestRenderAuditDiff_NewAndResolved(t *testing.T) {
	d := audit.Diff{
		New:      []audit.Fingerprint{{Gate: "size", Raw: "foo.go: 140 lines"}},
		Resolved: []audit.Fingerprint{{Gate: "size", Raw: "bar.go: 90 lines"}},
	}
	out := RenderAuditDiff(d)
	for _, want := range []string{
		"1 new", "1 resolved", "New (blocking)", "foo.go: 140 lines",
		"Resolved (prunable on next baseline)", "bar.go: 90 lines",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
	if strings.Contains(out, "no new violations") {
		t.Fatal("should not render the clean line when new violations exist")
	}
}

// TestRenderAuditDiff_NoNew covers the empty-section path and the clean
// "no new violations" line emitted when HasNew is false.
func TestRenderAuditDiff_NoNew(t *testing.T) {
	out := RenderAuditDiff(audit.Diff{})
	if !strings.Contains(out, "0 new") {
		t.Fatalf("expected zero-new header: %q", out)
	}
	if !strings.Contains(out, "no new violations since baseline") {
		t.Fatalf("expected clean line: %q", out)
	}
}
