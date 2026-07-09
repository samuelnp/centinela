package roadmap

import (
	"bytes"
	"strings"
	"testing"
)

// addBody: two schedulable phases followed by Backlog (the default insert anchor).
const addBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"a"}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"b"}]},` +
	`{"name":"Backlog","features":[]}]}`

// TestPhaseAdd_Positions covers --after (first/middle/before-Backlog anchors), no
// --after (before Backlog), no-Backlog (last), and --after Backlog (normal phase).
func TestPhaseAdd_Positions(t *testing.T) {
	cases := []struct {
		after string
		body  string
		want  string
	}{
		{"Phase 1: Foundations", addBody, "Phase 1: Foundations,X,Phase 2: Growth,Backlog"},
		{"Phase 2: Growth", addBody, "Phase 1: Foundations,Phase 2: Growth,X,Backlog"},
		{"", addBody, "Phase 1: Foundations,Phase 2: Growth,X,Backlog"},
		{"", `{"phases":[{"name":"Phase 1: Foundations","features":[]},{"name":"Phase 2: Growth","features":[]}]}`,
			"Phase 1: Foundations,Phase 2: Growth,X"},
		{"Backlog", addBody, "Phase 1: Foundations,Phase 2: Growth,Backlog,X"},
	}
	for _, c := range cases {
		p, _ := canonRoadmap(t, c.body)
		if err := PhaseAdd(p, "X", "", c.after); err != nil {
			t.Fatalf("after=%q: %v", c.after, err)
		}
		if got := strings.Join(phaseOrderNames(t, p), ","); got != c.want {
			t.Fatalf("after=%q: got %q want %q", c.after, got, c.want)
		}
	}
}

// TestPhaseAdd_EmptyFeaturesAndNote: a new phase has "features": [] and --note lands.
func TestPhaseAdd_EmptyFeaturesAndNote(t *testing.T) {
	p, _ := canonRoadmap(t, addBody)
	if err := PhaseAdd(p, "Phase 3: Scale", "hardening", ""); err != nil {
		t.Fatalf("PhaseAdd: %v", err)
	}
	got := string(crudBytes(t, p))
	if !strings.Contains(got, `"name": "Phase 3: Scale"`) {
		t.Fatalf("phase missing: %s", got)
	}
	if !strings.Contains(got, `"note": "hardening"`) {
		t.Fatalf("note missing: %s", got)
	}
	if len(orderIn(t, p, "Phase 3: Scale")) != 0 {
		t.Fatal("new phase must have empty features")
	}
}

// TestPhaseAdd_OnEmptyRoadmap: {"phases":[]} + add → the first phase.
func TestPhaseAdd_OnEmptyRoadmap(t *testing.T) {
	p := crudWrite(t, `{"phases":[]}`)
	if err := PhaseAdd(p, "Phase 1: Foundations", "", ""); err != nil {
		t.Fatalf("PhaseAdd: %v", err)
	}
	if got := strings.Join(phaseOrderNames(t, p), ","); got != "Phase 1: Foundations" {
		t.Fatalf("got %q", got)
	}
}

// TestPhaseAdd_UntouchedByteIdentical: existing phases round-trip byte-identical.
func TestPhaseAdd_UntouchedByteIdentical(t *testing.T) {
	p, before := canonRoadmap(t, addBody)
	if err := PhaseAdd(p, "X", "", "Phase 1: Foundations"); err != nil {
		t.Fatalf("PhaseAdd: %v", err)
	}
	after := crudBytes(t, p)
	if !bytes.Contains(after, phaseSlice(t, before, "Phase 1: Foundations")) {
		t.Fatal("Phase 1 must be byte-identical")
	}
	if !bytes.Contains(after, phaseSlice(t, before, "Phase 2: Growth")) {
		t.Fatal("Phase 2 must be byte-identical")
	}
}
