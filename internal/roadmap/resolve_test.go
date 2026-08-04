package roadmap

import (
	"strings"
	"testing"
)

// backlogDoc builds a roadmap.json whose Backlog holds the given raw entries.
func backlogDoc(entries ...string) []byte {
	return []byte(`{"phases":[{"name":"Phase 13","features":[{"name":"real"}]},` +
		`{"name":"Backlog","features":[` + strings.Join(entries, ",") + `]}]}`)
}

func finding(slug, at string) string {
	return `{"name":"` + slug + `","summary":"s","deferredAt":"` + at + `"}`
}

// AC12: both sides' appends survive, ordered by deferredAt then name.
func TestResolveUnionsBothSidesOfTheBacklog(t *testing.T) {
	base := backlogDoc(finding("b1", "2026-01-01T00:00:00Z"), finding("b2", "2026-01-02T00:00:00Z"))
	ours := backlogDoc(finding("b1", "2026-01-01T00:00:00Z"), finding("b2", "2026-01-02T00:00:00Z"),
		finding("o1", "2026-03-01T00:00:00Z"))
	theirs := backlogDoc(finding("b1", "2026-01-01T00:00:00Z"), finding("b2", "2026-01-02T00:00:00Z"),
		finding("t1", "2026-02-01T00:00:00Z"), finding("t2", "2026-02-02T00:00:00Z"))
	got, err := Resolve(base, ours, theirs)
	if err != nil {
		t.Fatal(err)
	}
	if got.Kept != 5 || got.FromBase != 2 || got.FromOurs != 1 || got.FromTheirs != 2 {
		t.Fatalf("arithmetic = %+v", got)
	}
	if got.FromBase+got.FromOurs+got.FromTheirs != got.Kept {
		t.Fatalf("the counts must reconcile to Kept: %+v", got)
	}
	doc := string(got.Doc)
	for _, slug := range []string{"b1", "b2", "o1", "t1", "t2"} {
		if !strings.Contains(doc, `"`+slug+`"`) {
			t.Fatalf("%s was dropped:\n%s", slug, doc)
		}
	}
	if strings.Index(doc, `"t1"`) > strings.Index(doc, `"o1"`) {
		t.Fatalf("findings must be ordered by deferredAt:\n%s", doc)
	}
}

// E11: the same slug on both sides survives once, keeping the first capture.
func TestResolveDedupesTheSameSlug(t *testing.T) {
	base := backlogDoc()
	ours := backlogDoc(`{"name":"same-thing","summary":"ours","deferredAt":"2026-05-02T00:00:00Z"}`)
	theirs := backlogDoc(`{"name":"same-thing","summary":"theirs","deferredAt":"2026-05-01T00:00:00Z"}`)
	got, err := Resolve(base, ours, theirs)
	if err != nil {
		t.Fatal(err)
	}
	if got.Kept != 1 || strings.Count(string(got.Doc), `"same-thing"`) != 1 {
		t.Fatalf("want exactly one entry, got %+v\n%s", got, got.Doc)
	}
	if !strings.Contains(string(got.Doc), "2026-05-01T00:00:00Z") {
		t.Fatalf("the earlier deferredAt must win:\n%s", got.Doc)
	}
	if got.FromOurs+got.FromTheirs != 1 || got.FromBase != 0 {
		t.Fatalf("a slug added on both sides must be counted once: %+v", got)
	}
	// The earlier capture kept is THEIRS, so theirs must be credited.
	if got.FromTheirs != 1 {
		t.Fatalf("the surviving entry is theirs; it must not be credited to ours: %+v", got)
	}
}

// E12: a one-sided deletion (promote/remove) wins over an untouched side.
func TestResolveHonoursAOneSidedDeletion(t *testing.T) {
	kept := finding("stays", "2026-01-01T00:00:00Z")
	gone := finding("promoted-away", "2026-01-02T00:00:00Z")
	got, err := Resolve(backlogDoc(kept, gone), backlogDoc(kept), backlogDoc(kept, gone))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got.Doc), "promoted-away") {
		t.Fatalf("a one-sided deletion must win:\n%s", got.Doc)
	}
	if !strings.Contains(string(got.Doc), "stays") || got.Kept != 1 {
		t.Fatalf("the untouched finding must survive: %+v\n%s", got, got.Doc)
	}
}
