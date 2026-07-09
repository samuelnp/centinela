package roadmap

import (
	"bytes"
	"testing"
)

// TestPhaseAdd_Refusals: duplicate/reserved(Backlog+Baseline)/empty/unknown-anchor,
// each leaving the roadmap byte-identical.
func TestPhaseAdd_Refusals(t *testing.T) {
	cases := []struct{ name, note, after, sub string }{
		{"Phase 1: Foundations", "", "", "already exists"},
		{"Backlog", "", "", "reserved phase name"},
		{"Baseline", "", "", "reserved phase name"},
		{"", "", "", "phase name is required"},
		{"Phase 3: Scale", "", "Phase 9: Nope", "unknown phase"},
	}
	for _, c := range cases {
		p, before := canonRoadmap(t, addBody)
		wantErr(t, PhaseAdd(p, c.name, c.note, c.after), c.sub)
		if !bytes.Equal(before, crudBytes(t, p)) {
			t.Fatalf("refused add %q must be byte-identical", c.name)
		}
	}
}
