package roadmap

import (
	"encoding/json"
	"strings"
	"testing"
)

// A malformed features array on one side is refused, never merged blind.
func TestResolveRefusesAMalformedFeaturesArray(t *testing.T) {
	good := backlogDoc(finding("a", "2026-01-01T00:00:00Z"))
	bad := []byte(`{"phases":[{"name":"Backlog","features":{"not":"an array"}}]}`)
	if _, err := Resolve(good, bad, good); err == nil {
		t.Fatal("a non-array features must be refused")
	}
	nameless := []byte(`{"phases":[{"name":"Backlog","features":[{"summary":"no name"}]}]}`)
	if _, err := Resolve(good, nameless, good); err == nil {
		t.Fatal("a finding with no name must be refused, not silently keyed on \"\"")
	}
}

// A phase entry that is not an object cannot be named, so the side is refused.
func TestResolveRefusesANonObjectPhase(t *testing.T) {
	good := backlogDoc()
	if _, err := Resolve(good, []byte(`{"phases":["oops"]}`), good); err == nil ||
		!strings.Contains(err.Error(), "our side") {
		t.Fatalf("want a named parse refusal, got %v", err)
	}
}

// Same deferredAt falls back to the name, so the order is total.
func TestBacklogOrderTieBreaksOnName(t *testing.T) {
	at := "2026-01-01T00:00:00Z"
	ours := backlogDoc(finding("zeta", at), finding("alpha", at))
	got, err := Resolve(backlogDoc(), ours, backlogDoc())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(string(got.Doc), `"alpha"`) > strings.Index(string(got.Doc), `"zeta"`) {
		t.Fatalf("an equal deferredAt must tie-break on name:\n%s", got.Doc)
	}
	if !byDeferredAtThenName(json.RawMessage(`{"name":"a"}`), json.RawMessage(`{"name":"b"}`)) {
		t.Fatal("two clockless entries must still order by name")
	}
}

// The Backlog phase existing only on their side must still be rebuilt, with its
// other keys carried over.
func TestResolveRebuildsABacklogOnlyOnTheirSide(t *testing.T) {
	base := []byte(`{"phases":[{"name":"P","features":[]}]}`)
	theirs := []byte(`{"phases":[{"name":"P","features":[]},` +
		`{"name":"Backlog","note":"deferred findings","features":[` + finding("t", "2026-01-01T00:00:00Z") + `]}]}`)
	got, err := Resolve(base, base, theirs)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"Backlog"`, `"deferred findings"`, `"t"`} {
		if !strings.Contains(string(got.Doc), want) {
			t.Fatalf("%s missing:\n%s", want, got.Doc)
		}
	}
	if got.Kept != 1 || got.FromTheirs != 1 {
		t.Fatalf("arithmetic = %+v", got)
	}
}

// Both sides emptied the Backlog: the phase survives with an empty array.
func TestResolveKeepsAnEmptiedBacklogPhase(t *testing.T) {
	base := backlogDoc(finding("gone", "2026-01-01T00:00:00Z"))
	empty := backlogDoc()
	got, err := Resolve(base, empty, empty)
	if err != nil {
		t.Fatal(err)
	}
	if got.Kept != 0 || !strings.Contains(string(got.Doc), `"features": []`) {
		t.Fatalf("want an empty Backlog phase: %+v\n%s", got, got.Doc)
	}
}

// The Backlog phase carrying no features key at all must not panic.
func TestResolveHandlesABacklogWithNoFeaturesKey(t *testing.T) {
	doc := []byte(`{"phases":[{"name":"Backlog"}]}`)
	got, err := Resolve(doc, doc, doc)
	if err != nil || got.Kept != 0 {
		t.Fatalf("err=%v got=%+v", err, got)
	}
}
