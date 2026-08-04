package roadmap

import (
	"errors"
	"strings"
	"testing"
)

const origEntry = `{"name":"finding-A","summary":"orig","deferredAt":"2026-01-01T00:00:00Z"}`
const theirEdit = `{"name":"finding-A","summary":"EDITED-BY-THEIRS","deferredAt":"2026-01-01T00:00:00Z"}`
const ourEdit = `{"name":"finding-A","summary":"EDITED-BY-OURS","deferredAt":"2026-01-01T00:00:00Z"}`

// THE round-2 regression: an edit does not move deferredAt, so ordering by
// deferredAt tied and silently returned OURS — discarding an incoming edit
// behind a green "✓ Resolved … kept N findings". A one-sided edit must win over
// an untouched side, whichever side made it.
func TestResolveKeepsAOneSidedEditFromTheirs(t *testing.T) {
	got, err := Resolve(backlogDoc(origEntry), backlogDoc(origEntry), backlogDoc(theirEdit))
	if err != nil {
		t.Fatalf("a one-sided incoming edit is not a conflict: %v", err)
	}
	if !strings.Contains(string(got.Doc), "EDITED-BY-THEIRS") {
		t.Fatalf("their edit was discarded:\n%s", got.Doc)
	}
	if strings.Contains(string(got.Doc), `"orig"`) {
		t.Fatalf("the base version must not survive an incoming edit:\n%s", got.Doc)
	}
	if got.Kept != 1 {
		t.Fatalf("Kept = %d, want 1", got.Kept)
	}
}

// The mirror direction must keep working.
func TestResolveKeepsAOneSidedEditFromOurs(t *testing.T) {
	got, err := Resolve(backlogDoc(origEntry), backlogDoc(ourEdit), backlogDoc(origEntry))
	if err != nil {
		t.Fatalf("a one-sided local edit is not a conflict: %v", err)
	}
	if !strings.Contains(string(got.Doc), "EDITED-BY-OURS") {
		t.Fatalf("our edit was discarded:\n%s", got.Doc)
	}
}

// Both sides editing the same finding differently is a real divergence: refuse.
func TestResolveRefusesABothSidesFindingEdit(t *testing.T) {
	_, err := Resolve(backlogDoc(origEntry), backlogDoc(ourEdit), backlogDoc(theirEdit))
	if err == nil {
		t.Fatal("two different edits to one finding must not be auto-merged")
	}
	var conflict *FindingConflictError
	if !errors.As(err, &conflict) || conflict.Slug != "finding-A" {
		t.Fatalf("want a *FindingConflictError naming finding-A, got %T %v", err, err)
	}
	if !strings.Contains(err.Error(), "modified differently on both sides") {
		t.Fatalf("the message must name the shape: %v", err)
	}
}

// Identical edits on both sides are agreement, not a conflict.
func TestResolveAcceptsIdenticalEditsOnBothSides(t *testing.T) {
	got, err := Resolve(backlogDoc(origEntry), backlogDoc(theirEdit), backlogDoc(theirEdit))
	if err != nil || !strings.Contains(string(got.Doc), "EDITED-BY-THEIRS") {
		t.Fatalf("agreement must merge cleanly: %v\n%s", err, got.Doc)
	}
}

// earlier() must no longer arbitrate anything the base had; it only decides two
// independent captures of one slug. This is the dedupe case, still working.
func TestEarlierStillDecidesIndependentCaptures(t *testing.T) {
	ours := `{"name":"same","summary":"ours","deferredAt":"2026-05-02T00:00:00Z"}`
	theirs := `{"name":"same","summary":"theirs","deferredAt":"2026-05-01T00:00:00Z"}`
	got, err := Resolve(backlogDoc(), backlogDoc(ours), backlogDoc(theirs))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got.Doc), "2026-05-01") || got.Kept != 1 {
		t.Fatalf("the earlier capture must win when the base never had it:\n%s", got.Doc)
	}
}
