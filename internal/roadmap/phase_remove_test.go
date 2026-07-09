package roadmap

import (
	"bytes"
	"strings"
	"testing"
)

// rmBody: an empty Phase 3, a non-empty Phase 2, plus a Backlog.
const rmBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"a"}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"billing-api"},{"name":"reporting"}]},` +
	`{"name":"Phase 3: Scale","features":[]},` +
	`{"name":"Backlog","features":[]}]}`

// TestPhaseRemove_EmptyPhase removes an empty phase; others stay byte-identical.
func TestPhaseRemove_EmptyPhase(t *testing.T) {
	p, before := canonRoadmap(t, rmBody)
	if err := PhaseRemove(p, "Phase 3: Scale", false); err != nil {
		t.Fatalf("PhaseRemove: %v", err)
	}
	if got := strings.Join(phaseOrderNames(t, p), ","); got != "Phase 1: Foundations,Phase 2: Growth,Backlog" {
		t.Fatalf("order wrong: %s", got)
	}
	if !bytes.Contains(crudBytes(t, p), phaseSlice(t, before, "Phase 1: Foundations")) {
		t.Fatal("Phase 1 must be byte-identical")
	}
}

// TestPhaseRemove_NonEmptyRefused: no --force refuses, naming the feature count,
// byte-identical.
func TestPhaseRemove_NonEmptyRefused(t *testing.T) {
	p, before := canonRoadmap(t, rmBody)
	wantErr(t, PhaseRemove(p, "Phase 2: Growth", false), "2 features")
	if !bytes.Equal(before, crudBytes(t, p)) {
		t.Fatal("refused remove must be byte-identical")
	}
}

// TestPhaseRemove_UnknownAndReserved: unknown → "not found"; Backlog/Baseline with
// or without --force → "reserved phase name"; each byte-identical.
func TestPhaseRemove_UnknownAndReserved(t *testing.T) {
	cases := []struct {
		name  string
		force bool
		sub   string
	}{
		{"Phase 9: Nope", false, "not found"},
		{"Backlog", false, "reserved phase name"},
		{"Backlog", true, "reserved phase name"},
		{"Baseline", true, "reserved phase name"},
	}
	for _, c := range cases {
		p, before := canonRoadmap(t, rmBody)
		wantErr(t, PhaseRemove(p, c.name, c.force), c.sub)
		if !bytes.Equal(before, crudBytes(t, p)) {
			t.Fatalf("refused remove %q must be byte-identical", c.name)
		}
	}
}

// TestPhaseRemove_OnlyPhase: removing the only phase yields exactly {"phases":[]}.
func TestPhaseRemove_OnlyPhase(t *testing.T) {
	p := crudWrite(t, `{"phases":[{"name":"Phase 1: Foundations","features":[]}]}`)
	if err := PhaseRemove(p, "Phase 1: Foundations", false); err != nil {
		t.Fatalf("PhaseRemove: %v", err)
	}
	if names := phaseOrderNames(t, p); len(names) != 0 {
		t.Fatalf("only-phase remove must leave zero phases, got %v", names)
	}
	got := strings.TrimSpace(string(crudBytes(t, p)))
	if strings.Contains(got, `"name"`) {
		t.Fatalf("empty roadmap must contain no phase, got %q", got)
	}
}
