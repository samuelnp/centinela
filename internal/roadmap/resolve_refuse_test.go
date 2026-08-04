package roadmap

import (
	"strings"
	"testing"
)

// AC13's headline: a real divergence must NOT be auto-merged.
func TestResolveRefusesABothSidesPhaseEdit(t *testing.T) {
	base := []byte(`{"phases":[{"name":"Phase 13: Lighter Centinela","features":[{"name":"a"}]}]}`)
	ours := []byte(`{"phases":[{"name":"Phase 13: Lighter Centinela","features":[{"name":"a"},{"name":"ours"}]}]}`)
	theirs := []byte(`{"phases":[{"name":"Phase 13: Lighter Centinela","features":[{"name":"a"},{"name":"theirs"}]}]}`)
	_, err := Resolve(base, ours, theirs)
	if err == nil || !strings.Contains(err.Error(), "Phase 13: Lighter Centinela") {
		t.Fatalf("want a named phase conflict, got %v", err)
	}
	var conflict *PhaseConflictError
	if !asPhaseConflict(err, &conflict) {
		t.Fatalf("want *PhaseConflictError, got %T", err)
	}
}

// E14: an unparseable side is refused BY NAME, never guessed at.
func TestResolveRefusesAnUnparseableSide(t *testing.T) {
	good := backlogDoc(finding("a", "2026-01-01T00:00:00Z"))
	if _, err := Resolve(good, good, []byte("{not json")); err == nil ||
		!strings.Contains(err.Error(), "their side") {
		t.Fatalf("want a named their-side parse error, got %v", err)
	}
	if _, err := Resolve(good, []byte("{{{"), good); err == nil ||
		!strings.Contains(err.Error(), "our side") {
		t.Fatalf("want a named our-side parse error, got %v", err)
	}
	if _, err := Resolve([]byte("nope"), good, good); err == nil ||
		!strings.Contains(err.Error(), "the base") {
		t.Fatalf("want a named base parse error, got %v", err)
	}
}
