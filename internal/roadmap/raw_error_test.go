package roadmap

import "testing"

// badNameBody has a feature whose name is the wrong JSON type, so featureName
// fails when the anchor/order helpers scan it.
const badNameBody = `{"phases":[{"name":"Phase 1: Foundations","features":[{"name":123}]}]}`

// TestAnchorPos_FeatureNameError covers anchorPos' featureName error branch.
func TestAnchorPos_FeatureNameError(t *testing.T) {
	doc := docFrom(t, badNameBody)
	if _, err := doc.anchorPos(0, "x", ""); err == nil {
		t.Fatal("malformed feature name must fail anchorPos")
	}
}

// TestPhaseOrder_FeatureNameError covers phaseOrder's featureName error branch.
func TestPhaseOrder_FeatureNameError(t *testing.T) {
	doc := docFrom(t, badNameBody)
	if _, err := doc.phaseOrder(); err == nil {
		t.Fatal("malformed feature name must fail phaseOrder")
	}
}

// TestSchedulablePhaseIndex_DecodeError covers schedulablePhaseIndex's decodePhase
// error branch: a well-formed source phase but a malformed later phase.
func TestSchedulablePhaseIndex_DecodeError(t *testing.T) {
	body := `{"phases":[` +
		`{"name":"Phase 1: Foundations","features":[{"name":"good"}]},` +
		`{"name":"Phase 2: Growth","features":"bad"}]}`
	p := crudWrite(t, body)
	if err := Move(p, MoveRequest{Slug: "good", ToPhase: "Phase 2: Growth"}); err == nil {
		t.Fatal("malformed target phase must fail schedulablePhaseIndex")
	}
}

// TestEdit_RenameNameScanError covers applyRename's phaseFeatureNames error branch
// via a malformed feature name in another slot during a rename.
func TestEdit_RenameNameScanError(t *testing.T) {
	body := `{"phases":[{"name":"Phase 1: Foundations","features":[{"name":"a"},{"name":123}]}]}`
	p := crudWrite(t, body)
	if err := Edit(p, EditRequest{Slug: "a", NewName: "a2"}); err == nil {
		t.Fatal("malformed sibling name must fail the collision scan")
	}
}

// TestReorder_PhaseOrderDecodeError covers phaseOrder's decodePhase error branch:
// the source phase is well-formed but a later phase is malformed.
func TestReorder_PhaseOrderDecodeError(t *testing.T) {
	body := `{"phases":[` +
		`{"name":"Phase 1: Foundations","features":[{"name":"b"},{"name":"a"}]},` +
		`{"name":"Phase 2: Growth","features":"bad"}]}`
	p := crudWrite(t, body)
	if err := Reorder(p, ReorderRequest{Slug: "b", BeforeAnchor: "a"}); err == nil {
		t.Fatal("malformed later phase must fail phaseOrder")
	}
}

// TestReorder_SelfAnchorNoop: anchoring a feature to itself is a no-op — Reorder
// exits 0 and leaves roadmap.json byte-identical (the self-anchor guard shared
// with Move short-circuits before any removal).
func TestReorder_SelfAnchorNoop(t *testing.T) {
	p, before := canonRoadmap(t, reorderBody)
	if err := Reorder(p, ReorderRequest{Slug: "checkout-ui", AfterAnchor: "checkout-ui"}); err != nil {
		t.Fatalf("self-anchor reorder must be a no-op, got %v", err)
	}
	if string(before) != string(crudBytes(t, p)) {
		t.Fatal("self-anchor reorder must be byte-identical")
	}
}
