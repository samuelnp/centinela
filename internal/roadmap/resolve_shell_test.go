package roadmap

import (
	"errors"
	"strings"
	"testing"
)

// notedBacklog builds a roadmap whose Backlog phase carries a `note` — a
// first-class Phase field rendered as the phase blockquote in ROADMAP.md.
func notedBacklog(note string, entries ...string) []byte {
	return []byte(`{"phases":[{"name":"Phase 13","features":[{"name":"real"}]},` +
		`{"name":"Backlog","note":"` + note + `","features":[` +
		strings.Join(entries, ",") + `]}]}`)
}

const shellEntry = `{"name":"b1","summary":"s","deferredAt":"2026-01-01T00:00:00Z"}`
const shellAdded = `{"name":"ours-new","summary":"s","deferredAt":"2026-02-01T00:00:00Z"}`

// THE round-3 regression: the Backlog phase SHELL was merged ours-wins instead
// of three-way, so an incoming edit to `note` vanished behind a green success
// line while the findings merged correctly.
func TestResolveKeepsAOneSidedShellEditFromTheirs(t *testing.T) {
	base := notedBacklog("ORIGINAL-NOTE", shellEntry)
	ours := notedBacklog("ORIGINAL-NOTE", shellEntry, shellAdded)
	theirs := notedBacklog("THEIRS-EDITED-NOTE", shellEntry)

	got, err := Resolve(base, ours, theirs)
	if err != nil {
		t.Fatalf("a one-sided shell edit is not a conflict: %v", err)
	}
	if !strings.Contains(string(got.Doc), "THEIRS-EDITED-NOTE") {
		t.Fatalf("their note edit was silently dropped:\n%s", got.Doc)
	}
	if strings.Contains(string(got.Doc), "ORIGINAL-NOTE") {
		t.Fatalf("the base note must not survive an incoming edit:\n%s", got.Doc)
	}
	// The findings must be unaffected by the shell fix.
	for _, slug := range []string{"b1", "ours-new"} {
		if !strings.Contains(string(got.Doc), `"`+slug+`"`) {
			t.Fatalf("%q was dropped:\n%s", slug, got.Doc)
		}
	}
}

// The mirror direction must keep working.
func TestResolveKeepsAOneSidedShellEditFromOurs(t *testing.T) {
	base := notedBacklog("ORIGINAL-NOTE", shellEntry)
	ours := notedBacklog("OURS-EDITED-NOTE", shellEntry)
	theirs := notedBacklog("ORIGINAL-NOTE", shellEntry, shellAdded)

	got, err := Resolve(base, ours, theirs)
	if err != nil {
		t.Fatalf("a one-sided local shell edit is not a conflict: %v", err)
	}
	if !strings.Contains(string(got.Doc), "OURS-EDITED-NOTE") {
		t.Fatalf("our note edit was dropped:\n%s", got.Doc)
	}
}

// Both sides editing the shell differently is a real divergence: refuse.
func TestResolveRefusesABothSidesShellEdit(t *testing.T) {
	base := notedBacklog("ORIGINAL-NOTE", shellEntry)
	ours := notedBacklog("OURS-EDITED-NOTE", shellEntry)
	theirs := notedBacklog("THEIRS-EDITED-NOTE", shellEntry)

	_, err := Resolve(base, ours, theirs)
	if err == nil {
		t.Fatal("two different note edits must not be auto-merged")
	}
	var conflict *PhaseConflictError
	if !errors.As(err, &conflict) || conflict.Phase != "Backlog" {
		t.Fatalf("want a *PhaseConflictError naming Backlog, got %T %v", err, err)
	}
}

// An unknown shell key added by one side must survive, and `features` must
// never be taken from a side — it is always the merged finding list.
func TestResolveShellKeepsUnknownKeysAndOwnsFeatures(t *testing.T) {
	base := []byte(`{"phases":[{"name":"Backlog","features":[` + shellEntry + `]}]}`)
	ours := []byte(`{"phases":[{"name":"Backlog","owner":"team-a","features":[` + shellEntry + `]}]}`)
	theirs := []byte(`{"phases":[{"name":"Backlog","features":[` + shellEntry + `,` + shellAdded + `]}]}`)

	got, err := Resolve(base, ours, theirs)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got.Doc), `"owner": "team-a"`) {
		t.Fatalf("a one-sided unknown shell key must survive:\n%s", got.Doc)
	}
	if got.Kept != 2 || !strings.Contains(string(got.Doc), "ours-new") {
		t.Fatalf("features must come from the merged list: %+v\n%s", got, got.Doc)
	}
}
