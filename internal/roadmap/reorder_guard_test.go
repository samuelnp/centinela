package roadmap

import (
	"bytes"
	"strings"
	"testing"
)

// reorderGuardBody adds a Backlog anchor to exercise the non-schedulable refusal.
const reorderGuardBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"auth-service"},{"name":"checkout-ui"}]},` +
	`{"name":"Backlog","features":[{"name":"legacy-finding","summary":"s"}]}]}`

// TestReorder_Refusals covers a Backlog anchor, an unknown slug, an unknown
// anchor, and a missing --before/--after — each byte-identical.
func TestReorder_Refusals(t *testing.T) {
	rows := []struct {
		name, slug, before, after, substr string
	}{
		{"backlog-anchor", "checkout-ui", "", "legacy-finding", "non-schedulable"},
		{"not-found", "ghost-feature", "auth-service", "", "not found"},
		{"unknown-anchor", "checkout-ui", "ghost-anchor", "", "not found"},
		{"no-anchor", "checkout-ui", "", "", "requires --before or --after"},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			p, before := canonRoadmap(t, reorderGuardBody)
			err := Reorder(p, ReorderRequest{Slug: r.slug, BeforeAnchor: r.before, AfterAnchor: r.after})
			if err == nil || !strings.Contains(err.Error(), r.substr) {
				t.Fatalf("want error %q, got %v", r.substr, err)
			}
			if !bytes.Equal(before, crudBytes(t, p)) {
				t.Fatal("refused reorder must be byte-identical")
			}
		})
	}
}
