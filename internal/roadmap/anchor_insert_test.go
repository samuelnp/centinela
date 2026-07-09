package roadmap

import (
	"strings"
	"testing"
)

// anchorBody: a single phase with three ordered features for index math.
const anchorBody = `{"phases":[{"name":"Phase 1: Foundations","features":[` +
	`{"name":"one"},{"name":"two"},{"name":"three"}]}]}`

// TestAnchorPos resolves before/after/append and rejects an unknown anchor.
func TestAnchorPos(t *testing.T) {
	doc := docFrom(t, anchorBody)
	cases := []struct {
		before, after string
		want          int
	}{
		{"one", "", 0},   // before first
		{"three", "", 2}, // before last
		{"", "one", 1},   // after first
		{"", "three", 3}, // after last
		{"two", "", 1},   // before middle
		{"", "", 3},      // no anchor => append at end
	}
	for _, c := range cases {
		got, err := doc.anchorPos(0, c.before, c.after)
		if err != nil || got != c.want {
			t.Fatalf("anchorPos(%q,%q)=%d,%v want %d", c.before, c.after, got, err, c.want)
		}
	}
	if _, err := doc.anchorPos(0, "ghost", ""); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("unknown anchor must error, got %v", err)
	}
}

// TestInsertFeatureAt inserts at the head, the tail, and rejects out-of-range.
func TestInsertFeatureAt(t *testing.T) {
	doc := docFrom(t, anchorBody)
	head, _ := compactBytes(Feature{Name: "head"})
	if err := doc.insertFeatureAt(0, 0, head); err != nil {
		t.Fatalf("insert head: %v", err)
	}
	order, _ := doc.phaseOrder()
	if strings.Join(order[0], ",") != "head,one,two,three" {
		t.Fatalf("head insert order: %v", order[0])
	}
	tail, _ := compactBytes(Feature{Name: "tail"})
	if err := doc.insertFeatureAt(0, 4, tail); err != nil {
		t.Fatalf("insert tail: %v", err)
	}
	order, _ = doc.phaseOrder()
	if last := order[0][len(order[0])-1]; last != "tail" {
		t.Fatalf("tail insert order: %v", order[0])
	}
	if err := doc.insertFeatureAt(0, 99, tail); err == nil {
		t.Fatal("out-of-range insert position must error")
	}
}
