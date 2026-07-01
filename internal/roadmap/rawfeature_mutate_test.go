package roadmap

import (
	"strings"
	"testing"
)

// renderStr renders a doc to a string, failing on error.
func renderStr(t *testing.T, d *rawDoc) string {
	t.Helper()
	b, err := d.render()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return string(b)
}

// TestAppendFeatureToPhase appends to a schedulable phase and renders stably.
func TestAppendFeatureToPhase(t *testing.T) {
	doc := docFrom(t, crudBody)
	entry, _ := compactBytes(Feature{Name: "widget", Draft: true})
	if err := doc.appendFeatureToPhase("Phase 1: Foundations", entry); err != nil {
		t.Fatalf("append: %v", err)
	}
	out := renderStr(t, doc)
	if !strings.Contains(out, `{"name":"widget","draft":true}`) {
		t.Fatalf("appended entry must render one-per-line: %s", out)
	}
	// Untouched Phase 2 round-trips byte-stably; re-render is identical.
	if out != renderStr(t, doc) {
		t.Fatal("render must be deterministic")
	}
}

// TestAppendFeatureToPhase_Rejections covers duplicate and unknown/non-schedulable.
func TestAppendFeatureToPhase_Rejections(t *testing.T) {
	doc := docFrom(t, crudBody)
	dup, _ := compactBytes(Feature{Name: "auth-service"})
	if err := doc.appendFeatureToPhase("Phase 1: Foundations", dup); err == nil ||
		!strings.Contains(err.Error(), "already exists") {
		t.Fatalf("duplicate must be refused, got %v", err)
	}
	e, _ := compactBytes(Feature{Name: "x"})
	if err := doc.appendFeatureToPhase("Phase 9", e); err == nil ||
		!strings.Contains(err.Error(), "unknown phase") {
		t.Fatalf("unknown phase must error, got %v", err)
	}
	if err := doc.appendFeatureToPhase("Backlog", e); err == nil ||
		!strings.Contains(err.Error(), "unknown phase") {
		t.Fatalf("Backlog (non-schedulable) must error unknown phase, got %v", err)
	}
}

// TestRemoveFeatureAt drops a feature and leaves an empty array when it was last.
func TestRemoveFeatureAt(t *testing.T) {
	doc := docFrom(t, crudBody)
	if err := doc.removeFeatureAt(2, "lonely-feature"); err != nil {
		t.Fatalf("removeFeatureAt: %v", err)
	}
	out := renderStr(t, doc)
	if strings.Contains(out, "lonely-feature") {
		t.Fatal("feature must be removed")
	}
	if !strings.Contains(out, `"features": []`) {
		t.Fatalf("emptied phase must keep []: %s", out)
	}
}

// TestReplaceFeatureAt swaps an entry and rejects an out-of-range index.
func TestReplaceFeatureAt(t *testing.T) {
	doc := docFrom(t, crudBody)
	cleared, _ := compactBytes(Feature{Name: "billing-api"})
	if err := doc.replaceFeatureAt(1, 0, cleared); err != nil {
		t.Fatalf("replaceFeatureAt: %v", err)
	}
	if err := doc.replaceFeatureAt(1, 9, cleared); err == nil {
		t.Fatal("out-of-range featIdx must error")
	}
}
