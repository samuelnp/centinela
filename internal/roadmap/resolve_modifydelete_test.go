package roadmap

import (
	"errors"
	"strings"
	"testing"
)

// F4: ours EDITS a finding, theirs DELETES it. Deciding by presence alone threw
// the edit away with exit 0 and a `kept 0` summary — the exact silent loss this
// command exists to prevent. It must refuse, by name.
func TestResolveRefusesAModifyDeletePair(t *testing.T) {
	base := backlogDoc(`{"name":"finding-A","summary":"original","deferredAt":"2026-01-01T00:00:00Z"}`)
	ours := backlogDoc(`{"name":"finding-A","summary":"IMPORTANT-NEW-DETAIL","deferredAt":"2026-01-01T00:00:00Z"}`)
	theirs := backlogDoc()

	_, err := Resolve(base, ours, theirs)
	if err == nil {
		t.Fatal("an edited finding must not be dropped by the other side's delete")
	}
	var conflict *FindingConflictError
	if !errors.As(err, &conflict) || conflict.Slug != "finding-A" {
		t.Fatalf("want a *FindingConflictError naming finding-A, got %T %v", err, err)
	}
	if !strings.Contains(err.Error(), "finding-A") ||
		!strings.Contains(err.Error(), "modified on one side and deleted") {
		t.Fatalf("the message must name the finding and the shape: %v", err)
	}
}

// Symmetric: theirs edits, ours deletes.
func TestResolveRefusesAModifyDeletePairFromEitherSide(t *testing.T) {
	base := backlogDoc(`{"name":"finding-A","summary":"original","deferredAt":"2026-01-01T00:00:00Z"}`)
	edited := backlogDoc(`{"name":"finding-A","summary":"THEIR-EDIT","deferredAt":"2026-01-01T00:00:00Z"}`)
	if _, err := Resolve(base, backlogDoc(), edited); err == nil {
		t.Fatal("their edit must not be dropped by our delete either")
	}
}

// The unchanged case must keep working: a delete against an UNTOUCHED side wins.
func TestResolveStillHonoursDeleteAgainstAnUntouchedSide(t *testing.T) {
	entry := `{"name":"promoted-away","summary":"s","deferredAt":"2026-01-01T00:00:00Z"}`
	kept := `{"name":"stays","summary":"s","deferredAt":"2026-01-02T00:00:00Z"}`
	got, err := Resolve(backlogDoc(entry, kept), backlogDoc(kept), backlogDoc(entry, kept))
	if err != nil {
		t.Fatalf("an untouched side must not read as an edit: %v", err)
	}
	if strings.Contains(string(got.Doc), "promoted-away") || got.Kept != 1 {
		t.Fatalf("the deletion must still win: %+v\n%s", got, got.Doc)
	}
}

// Whitespace alone is not an edit: the compacted comparison must ignore it.
func TestResolveDeleteWinsOverAReformattedButUnchangedSide(t *testing.T) {
	entry := `{"name":"gone","summary":"s","deferredAt":"2026-01-01T00:00:00Z"}`
	spaced := `{ "name" : "gone" , "summary" : "s" , "deferredAt" : "2026-01-01T00:00:00Z" }`
	got, err := Resolve(backlogDoc(entry), backlogDoc(spaced), backlogDoc())
	if err != nil {
		t.Fatalf("a reformat is not an edit: %v", err)
	}
	if got.Kept != 0 {
		t.Fatalf("the deletion must win: %+v\n%s", got, got.Doc)
	}
}

// A refusal must reach the operator before anything is written or staged.
func TestResolveModifyDeleteRefusalWritesNothing(t *testing.T) {
	base := backlogDoc(`{"name":"a","summary":"one","deferredAt":"2026-01-01T00:00:00Z"}`)
	ours := backlogDoc(`{"name":"a","summary":"two","deferredAt":"2026-01-01T00:00:00Z"}`)
	got, err := Resolve(base, ours, backlogDoc())
	if err == nil {
		t.Fatal("want a refusal")
	}
	if got.Doc != nil || got.Kept != 0 {
		t.Fatalf("a refused merge must return no document: %+v", got)
	}
}
