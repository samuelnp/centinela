package roadmap

import (
	"bytes"
	"strings"
	"testing"
)

// moveBody: a lone source in Phase 1, a three-feature target in Phase 2, and an
// untouched Phase 3 to prove byte-identical preservation.
const moveBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"checkout-ui"}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"billing-api"},{"name":"reporting"},` +
	`{"name":"invoicing"}]},` +
	`{"name":"Phase 3: Solo","features":[{"name":"lonely-feature"}]}]}`

// TestMove_AppendsByDefault relocates to the target phase's end; untouched phases
// stay byte-identical.
func TestMove_AppendsByDefault(t *testing.T) {
	p, before := canonRoadmap(t, moveBody)
	if err := Move(p, MoveRequest{Slug: "checkout-ui", ToPhase: "Phase 2: Growth"}); err != nil {
		t.Fatalf("Move: %v", err)
	}
	if contains(orderIn(t, p, "Phase 1: Foundations"), "checkout-ui") {
		t.Fatal("checkout-ui must leave Phase 1")
	}
	got := orderIn(t, p, "Phase 2: Growth")
	if strings.Join(got, ",") != "billing-api,reporting,invoicing,checkout-ui" {
		t.Fatalf("append order wrong: %v", got)
	}
	if !bytes.Contains(crudBytes(t, p), phaseSlice(t, before, "Phase 3: Solo")) {
		t.Fatal("untouched Phase 3 must be byte-identical")
	}
}

// TestMove_Anchors covers first/last/middle placement via --before/--after.
func TestMove_Anchors(t *testing.T) {
	cases := []struct {
		before, after, want string
	}{
		{"billing-api", "", "checkout-ui,billing-api,reporting,invoicing"},
		{"", "invoicing", "billing-api,reporting,invoicing,checkout-ui"},
		{"", "billing-api", "billing-api,checkout-ui,reporting,invoicing"},
		{"invoicing", "", "billing-api,reporting,checkout-ui,invoicing"},
	}
	for _, c := range cases {
		p, _ := canonRoadmap(t, moveBody)
		req := MoveRequest{Slug: "checkout-ui", ToPhase: "Phase 2: Growth",
			BeforeAnchor: c.before, AfterAnchor: c.after}
		if err := Move(p, req); err != nil {
			t.Fatalf("Move %+v: %v", c, err)
		}
		if got := strings.Join(orderIn(t, p, "Phase 2: Growth"), ","); got != c.want {
			t.Fatalf("anchor %+v: got %q want %q", c, got, c.want)
		}
	}
}

// TestMove_SelfAnchorNoop: anchoring a feature to itself is a no-op — Move exits
// 0 and leaves roadmap.json byte-identical (no silent mutation), matching the
// spec's "move self-anchor no-op" scenario.
func TestMove_SelfAnchorNoop(t *testing.T) {
	body := `{"phases":[{"name":"Phase 1: Foundations","features":[{"name":"checkout-ui"}]}]}`
	p, before := canonRoadmap(t, body)
	if err := Move(p, MoveRequest{Slug: "checkout-ui", ToPhase: "Phase 1: Foundations", AfterAnchor: "checkout-ui"}); err != nil {
		t.Fatalf("self-anchor move must be a no-op, got %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, p)) {
		t.Fatal("self-anchor move must leave roadmap.json byte-identical")
	}
}
