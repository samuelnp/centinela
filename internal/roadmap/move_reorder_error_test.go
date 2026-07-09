package roadmap

import "testing"

// TestMove_FinalizeDecodeError drives finalizeMutation's toRoadmap error: a
// malformed sibling in the source phase decodes fine name-wise (so the move
// mechanics run) but fails the post-mutation typed re-decode.
func TestMove_FinalizeDecodeError(t *testing.T) {
	body := `{"phases":[` +
		`{"name":"Phase 1: Foundations","features":[{"name":"good"},{"name":"bad","dependsOn":"x"}]},` +
		`{"name":"Phase 2: Growth","features":[]}]}`
	p := crudWrite(t, body)
	if err := Move(p, MoveRequest{Slug: "good", ToPhase: "Phase 2: Growth"}); err == nil {
		t.Fatal("malformed sibling must fail the post-move re-decode")
	}
}

// TestReorder_FinalizeDecodeError drives the same path for reorder: a real
// reposition reaches finalizeMutation, where the malformed sibling fails decode.
func TestReorder_FinalizeDecodeError(t *testing.T) {
	body := `{"phases":[{"name":"Phase 1: Foundations","features":[` +
		`{"name":"a"},{"name":"b"},{"name":"bad","dependsOn":"x"}]}]}`
	p := crudWrite(t, body)
	if err := Reorder(p, ReorderRequest{Slug: "b", BeforeAnchor: "a"}); err == nil {
		t.Fatal("malformed sibling must fail the post-reorder re-decode")
	}
}

// TestMove_UnknownAnchorAfterRemove covers Move's anchorPos error path: a valid
// target with an anchor that names no sibling there.
func TestMove_UnknownAnchorAfterRemove(t *testing.T) {
	p, before := canonRoadmap(t, moveBody)
	err := Move(p, MoveRequest{Slug: "checkout-ui", ToPhase: "Phase 2: Growth", AfterAnchor: "ghost"})
	if err == nil {
		t.Fatal("unknown anchor must error")
	}
	if string(before) != string(crudBytes(t, p)) {
		t.Fatal("failed move must be byte-identical")
	}
}
