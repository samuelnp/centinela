package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/audit"
)

// TestRenderAdoptionWithFindings shows per-gate counts, the total, and the
// ratchet-to-zero framing.
func TestRenderAdoptionWithFindings(t *testing.T) {
	o := audit.Outcome{Path: ".workflow/audit-baseline.json", Baseline: audit.Baseline{
		Gates: []audit.GateEntry{
			{Gate: "G1: File Size", Fingerprints: audit.Compute("G1: File Size", []string{"a.go (150 lines)", "b.go (140 lines)"})},
		},
	}}
	out := RenderAdoption(o)
	for _, want := range []string{"G1: File Size", "2 accepted finding", "ratchet to zero"} {
		if !strings.Contains(out, want) {
			t.Fatalf("render missing %q:\n%s", want, out)
		}
	}
}

// TestRenderAdoptionZeroFindings states there is nothing to ratchet.
func TestRenderAdoptionZeroFindings(t *testing.T) {
	out := RenderAdoption(audit.Outcome{Path: ".workflow/audit-baseline.json"})
	if !strings.Contains(out, "0 accepted findings") || !strings.Contains(out, "nothing to ratchet") {
		t.Fatalf("zero-finding render unexpected:\n%s", out)
	}
}
