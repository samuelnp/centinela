package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/audit"
)

func TestRenderAuditDiff_ZeroValue(t *testing.T) {
	out := RenderAuditDiff(audit.Diff{})
	if !strings.Contains(out, "no new violations") {
		t.Fatalf("zero diff should emit no-new message, got %q", out)
	}
	if !strings.Contains(out, "0 new") {
		t.Fatalf("header should mention 0 new, got %q", out)
	}
}

func TestRenderAuditDiff_WithNew(t *testing.T) {
	d := audit.Diff{
		New:      []audit.Fingerprint{{Gate: "size", Raw: "big.go:130"}},
		Resolved: []audit.Fingerprint{{Gate: "size", Raw: "old.go:100"}},
	}
	out := RenderAuditDiff(d)
	if !strings.Contains(out, "New (blocking)") {
		t.Fatalf("want blocking section, got %q", out)
	}
	if !strings.Contains(out, "Resolved") {
		t.Fatalf("want resolved section, got %q", out)
	}
	if !strings.Contains(out, "big.go:130") {
		t.Fatalf("want raw detail in output, got %q", out)
	}
}

func TestRenderAuditDiff_OnlyResolved(t *testing.T) {
	d := audit.Diff{
		Resolved: []audit.Fingerprint{{Gate: "size", Raw: "gone.go:90"}},
	}
	out := RenderAuditDiff(d)
	if !strings.Contains(out, "no new violations") {
		t.Fatalf("no-new message should appear when resolved-only, got %q", out)
	}
	if !strings.Contains(out, "Resolved") {
		t.Fatalf("resolved section should appear, got %q", out)
	}
}

func TestRenderAuditDiff_Counts(t *testing.T) {
	d := audit.Diff{
		New:       []audit.Fingerprint{{Gate: "g", Raw: "a"}, {Gate: "g", Raw: "b"}},
		Baselined: []audit.Fingerprint{{Gate: "g", Raw: "c"}},
	}
	out := RenderAuditDiff(d)
	if !strings.Contains(out, "2 new") {
		t.Fatalf("want '2 new' in header, got %q", out)
	}
	if !strings.Contains(out, "1 baselined") {
		t.Fatalf("want '1 baselined' in header, got %q", out)
	}
}
