package roadmap

import (
	"bytes"
	"testing"
)

// TestEdit_NoopByteIdentical: an edit that changes nothing — a same-name rename or
// no fields at all — is a byte-identical no-op (the file is not rewritten), while
// a nonexistent slug still errors because the guard runs after findFeature.
func TestEdit_NoopByteIdentical(t *testing.T) {
	body := `{"phases":[{"name":"Phase 1: Foundations","features":[{"name":"checkout-ui","dependsOn":[]}]}]}`
	p, before := canonRoadmap(t, body)

	if err := Edit(p, EditRequest{Slug: "checkout-ui", NewName: "checkout-ui"}); err != nil {
		t.Fatalf("same-name edit must be a no-op, got %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, p)) {
		t.Fatal("same-name edit must leave roadmap.json byte-identical")
	}

	if err := Edit(p, EditRequest{Slug: "checkout-ui"}); err != nil {
		t.Fatalf("empty edit must be a no-op, got %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, p)) {
		t.Fatal("empty edit must leave roadmap.json byte-identical")
	}

	if err := Edit(p, EditRequest{Slug: "ghost"}); err == nil {
		t.Fatal("edit of a nonexistent feature must error (guard runs after findFeature)")
	}
}
