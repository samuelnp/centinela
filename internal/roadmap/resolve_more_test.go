package roadmap

import (
	"errors"
	"strings"
	"testing"
)

// asPhaseConflict is errors.As with the concrete type this package refuses on.
func asPhaseConflict(err error, target **PhaseConflictError) bool {
	return errors.As(err, target)
}

// A one-sided phase edit is not a conflict — it is the side that moved.
func TestResolveTakesTheSideThatMoved(t *testing.T) {
	base := []byte(`{"phases":[{"name":"P","features":[{"name":"a"}]}]}`)
	ours := []byte(`{"phases":[{"name":"P","features":[{"name":"a"},{"name":"new"}]}]}`)
	got, err := Resolve(base, ours, base)
	if err != nil || !strings.Contains(string(got.Doc), "new") {
		t.Fatalf("our one-sided edit must win: %v\n%s", err, got.Doc)
	}
	got, err = Resolve(base, base, ours)
	if err != nil || !strings.Contains(string(got.Doc), "new") {
		t.Fatalf("their one-sided edit must win: %v\n%s", err, got.Doc)
	}
}

// A phase only one side has must survive; a whole phase both sides added
// identically is agreement, not a conflict.
func TestResolveKeepsAOneSidedNewPhase(t *testing.T) {
	base := []byte(`{"phases":[{"name":"P","features":[]}]}`)
	theirs := []byte(`{"phases":[{"name":"P","features":[]},{"name":"Phase 14","features":[]}]}`)
	got, err := Resolve(base, base, theirs)
	if err != nil || !strings.Contains(string(got.Doc), "Phase 14") {
		t.Fatalf("a new phase must survive: %v\n%s", err, got.Doc)
	}
	if got, err = Resolve(base, theirs, theirs); err != nil ||
		strings.Count(string(got.Doc), "Phase 14") != 1 {
		t.Fatalf("identical additions are agreement: %v\n%s", err, got.Doc)
	}
}

// Formatting alone is never a semantic conflict.
func TestResolveIgnoresWhitespaceOnlyDifferences(t *testing.T) {
	base := []byte(`{"phases":[{"name":"P","features":[{"name":"a"}]}]}`)
	reformatted := []byte("{\n  \"phases\": [\n    {\n      \"name\": \"P\",\n" +
		"      \"features\": [\n        {\"name\":\"a\"}\n      ]\n    }\n  ]\n}\n")
	ours := []byte(`{"phases":[{"name":"P","features":[{"name":"a"},{"name":"new"}]}]}`)
	if _, err := Resolve(base, ours, reformatted); err != nil {
		t.Fatalf("a reformat must not read as a semantic edit: %v", err)
	}
}

// A one-sided intro edit survives; a both-sides one is refused by name.
func TestResolveThreeWaysTopLevelKeys(t *testing.T) {
	base := []byte(`{"intro":"a","phases":[]}`)
	ours := []byte(`{"intro":"b","phases":[]}`)
	theirs := []byte(`{"intro":"c","phases":[]}`)
	got, err := Resolve(base, ours, base)
	if err != nil || !strings.Contains(string(got.Doc), `"b"`) {
		t.Fatalf("a one-sided intro edit must survive: %v\n%s", err, got.Doc)
	}
	if _, err := Resolve(base, ours, theirs); err == nil || !strings.Contains(err.Error(), "intro") {
		t.Fatalf("want a named top-level conflict, got %v", err)
	}
}

// An absent stage (one side added or deleted the file) reads as empty.
func TestResolveHandlesAnAbsentStage(t *testing.T) {
	added := backlogDoc(finding("a", "2026-01-01T00:00:00Z"))
	got, err := Resolve(nil, added, added)
	if err != nil || got.Kept != 1 {
		t.Fatalf("an absent base must read as empty: %v %+v", err, got)
	}
	if got.FromOurs != 1 || got.FromTheirs != 0 {
		t.Fatalf("a both-sides addition counts once: %+v", got)
	}
}

// Backlog entries with no clock still round-trip, sorted first.
func TestResolveOrdersUnknownDeferredAtFirst(t *testing.T) {
	ours := backlogDoc(`{"name":"legacy","summary":"s"}`, finding("dated", "2026-01-01T00:00:00Z"))
	got, err := Resolve(backlogDoc(), ours, backlogDoc())
	if err != nil {
		t.Fatal(err)
	}
	doc := string(got.Doc)
	if strings.Index(doc, `"legacy"`) > strings.Index(doc, `"dated"`) {
		t.Fatalf("an entry with no deferredAt must sort first:\n%s", doc)
	}
}
