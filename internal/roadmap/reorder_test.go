package roadmap

import (
	"bytes"
	"strings"
	"testing"
)

// reorderBody: a three-feature Phase 1 for within-phase moves, plus Phase 2 for
// cross-phase reorders.
const reorderBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"auth-service"},` +
	`{"name":"checkout-ui"},{"name":"billing-api"}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"reporting"}]}]}`

// TestReorder_WithinPhase repositions a feature ahead of a same-phase anchor;
// the untouched Phase 2 stays byte-identical.
func TestReorder_WithinPhase(t *testing.T) {
	p, before := canonRoadmap(t, reorderBody)
	if err := Reorder(p, ReorderRequest{Slug: "billing-api", BeforeAnchor: "auth-service"}); err != nil {
		t.Fatalf("Reorder: %v", err)
	}
	if got := strings.Join(orderIn(t, p, "Phase 1: Foundations"), ","); got != "billing-api,auth-service,checkout-ui" {
		t.Fatalf("within-phase order wrong: %s", got)
	}
	if !bytes.Contains(crudBytes(t, p), phaseSlice(t, before, "Phase 2: Growth")) {
		t.Fatal("untouched Phase 2 must be byte-identical")
	}
}

// TestReorder_AcrossPhase moves a feature next to an anchor in another phase.
func TestReorder_AcrossPhase(t *testing.T) {
	p, _ := canonRoadmap(t, reorderBody)
	if err := Reorder(p, ReorderRequest{Slug: "checkout-ui", AfterAnchor: "reporting"}); err != nil {
		t.Fatalf("Reorder: %v", err)
	}
	if contains(orderIn(t, p, "Phase 1: Foundations"), "checkout-ui") {
		t.Fatal("checkout-ui must leave Phase 1")
	}
	if got := strings.Join(orderIn(t, p, "Phase 2: Growth"), ","); got != "reporting,checkout-ui" {
		t.Fatalf("cross-phase order wrong: %s", got)
	}
}

// TestReorder_NoOpByteIdentical asserts an order-preserving reorder does not write.
func TestReorder_NoOpByteIdentical(t *testing.T) {
	p, before := canonRoadmap(t, reorderBody)
	// checkout-ui is already immediately after auth-service.
	if err := Reorder(p, ReorderRequest{Slug: "checkout-ui", AfterAnchor: "auth-service"}); err != nil {
		t.Fatalf("Reorder: %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, p)) {
		t.Fatal("no-op reorder must leave roadmap.json byte-identical")
	}
}
