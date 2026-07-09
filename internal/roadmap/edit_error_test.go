package roadmap

import (
	"path/filepath"
	"testing"
)

// missing is a path with no roadmap.json, to drive the readRawRoadmap error path.
func missing(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "nope.json")
}

// TestEdit_MissingFile surfaces the read error before any mutation.
func TestEdit_MissingFile(t *testing.T) {
	if err := Edit(missing(t), EditRequest{Slug: "x", Description: "y"}); err == nil {
		t.Fatal("missing roadmap.json must error")
	}
}

// TestMove_MissingFile and TestReorder_MissingFile cover their read-error paths.
func TestMove_MissingFile(t *testing.T) {
	if err := Move(missing(t), MoveRequest{Slug: "x", ToPhase: "P"}); err == nil {
		t.Fatal("missing roadmap.json must error")
	}
}

func TestReorder_MissingFile(t *testing.T) {
	if err := Reorder(missing(t), ReorderRequest{Slug: "x", BeforeAnchor: "a"}); err == nil {
		t.Fatal("missing roadmap.json must error")
	}
}

// TestEdit_ArchetypeField applies the archetype-only branch of applyEditFields.
func TestEdit_ArchetypeField(t *testing.T) {
	p, _ := canonRoadmap(t, editBody)
	if err := Edit(p, EditRequest{Slug: "checkout-ui", Archetype: "spike"}); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if got := featureIn(t, p, "checkout-ui").Archetype; got != "spike" {
		t.Fatalf("archetype not applied: %q", got)
	}
}

// TestEdit_MalformedTargetDecode drives Edit's json.Unmarshal error branch: the
// target feature has a dependsOn of the wrong JSON type.
func TestEdit_MalformedTargetDecode(t *testing.T) {
	body := `{"phases":[{"name":"Phase 1: Foundations","features":[{"name":"bad","dependsOn":"x"}]}]}`
	p := crudWrite(t, body)
	if err := Edit(p, EditRequest{Slug: "bad", Description: "y"}); err == nil {
		t.Fatal("malformed target feature must fail typed decode")
	}
}

// TestEdit_RenameRewriteDecodeError drives rewriteDependents' decode error via a
// malformed sibling while renaming a well-formed feature.
func TestEdit_RenameRewriteDecodeError(t *testing.T) {
	body := `{"phases":[{"name":"Phase 1: Foundations","features":[{"name":"a"},` +
		`{"name":"bad","dependsOn":"x"}]}]}`
	p := crudWrite(t, body)
	if err := Edit(p, EditRequest{Slug: "a", NewName: "a2"}); err == nil {
		t.Fatal("malformed sibling must fail the dependent rewrite")
	}
}
