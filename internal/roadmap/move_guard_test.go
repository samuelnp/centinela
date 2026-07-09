package roadmap

import (
	"bytes"
	"strings"
	"testing"
)

// guardBody holds schedulable phases plus non-schedulable Backlog/Baseline, to
// exercise every move refusal path.
const guardBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"checkout-ui"}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"billing-api"}]},` +
	`{"name":"Backlog","features":[{"name":"legacy-finding","summary":"s"}]},` +
	`{"name":"Baseline","features":[{"name":"shipped-item","summary":"s"}]}]}`

// TestMove_Refusals runs every guard row and asserts each leaves roadmap.json
// byte-identical. Substrings match the implementation's wording (which uses
// "non-schedulable"/"not found" where the spec examples say "unknown phase/feature"
// — see qa-senior deferred findings).
func TestMove_Refusals(t *testing.T) {
	rows := []struct {
		name, slug, toPhase, before, substr string
	}{
		{"target-backlog", "checkout-ui", "Backlog", "", "non-schedulable"},
		{"target-baseline", "checkout-ui", "Baseline", "", "non-schedulable"},
		{"unknown-phase", "checkout-ui", "Phase 9: Nonexistent", "", "unknown phase"},
		{"unknown-anchor", "checkout-ui", "Phase 2: Growth", "ghost-anchor", "not found"},
		{"source-backlog", "legacy-finding", "Phase 2: Growth", "", "non-schedulable"},
		{"source-baseline", "shipped-item", "Phase 2: Growth", "", "non-schedulable"},
		{"not-found", "ghost-feature", "Phase 2: Growth", "", "not found"},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			p, before := canonRoadmap(t, guardBody)
			err := Move(p, MoveRequest{Slug: r.slug, ToPhase: r.toPhase, BeforeAnchor: r.before})
			if err == nil || !strings.Contains(err.Error(), r.substr) {
				t.Fatalf("want error %q, got %v", r.substr, err)
			}
			if !bytes.Equal(before, crudBytes(t, p)) {
				t.Fatal("refused move must be byte-identical")
			}
		})
	}
}

// TestMove_PreservesDraft moves a draft feature and keeps its draft flag verbatim.
func TestMove_PreservesDraft(t *testing.T) {
	body := `{"phases":[` +
		`{"name":"Phase 1: Foundations","features":[{"name":"new-widget","draft":true}]},` +
		`{"name":"Phase 2: Growth","features":[{"name":"billing-api"}]}]}`
	p, _ := canonRoadmap(t, body)
	if err := Move(p, MoveRequest{Slug: "new-widget", ToPhase: "Phase 2: Growth"}); err != nil {
		t.Fatalf("Move: %v", err)
	}
	f := featureIn(t, p, "new-widget")
	if !f.Draft {
		t.Fatalf("draft flag must survive the move: %+v", f)
	}
	if !contains(orderIn(t, p, "Phase 2: Growth"), "new-widget") {
		t.Fatal("new-widget must land in Phase 2")
	}
}
