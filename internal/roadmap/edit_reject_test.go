package roadmap

import (
	"bytes"
	"strings"
	"testing"
)

// TestEdit_UnknownDepByteIdentical rejects a dependsOn on a non-existent feature
// after the in-memory mutation, leaving the file byte-identical.
func TestEdit_UnknownDepByteIdentical(t *testing.T) {
	p, before := canonRoadmap(t, renameBody)
	err := Edit(p, EditRequest{Slug: "checkout-ui", DependsOn: []string{"ghost-feature"}, SetDeps: true})
	if err == nil || !strings.Contains(err.Error(), "depends on unknown feature") {
		t.Fatalf("unknown dep must error, got %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, p)) {
		t.Fatal("rejected edit must be byte-identical")
	}
}

// TestEdit_SelfCycleByteIdentical rejects a feature depending on itself.
func TestEdit_SelfCycleByteIdentical(t *testing.T) {
	p, before := canonRoadmap(t, renameBody)
	err := Edit(p, EditRequest{Slug: "checkout-ui", DependsOn: []string{"checkout-ui"}, SetDeps: true})
	if err == nil || !strings.Contains(err.Error(), "cycle detected") {
		t.Fatalf("self-cycle must error, got %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, p)) {
		t.Fatal("rejected edit must be byte-identical")
	}
}

// TestEdit_MultiHopCycleByteIdentical rejects a two-feature cycle introduced via
// --depends-on (auth-service already depends on checkout-ui).
func TestEdit_MultiHopCycleByteIdentical(t *testing.T) {
	body := `{"phases":[{"name":"Phase 1: Foundations","features":[` +
		`{"name":"auth-service","dependsOn":["checkout-ui"]},` +
		`{"name":"checkout-ui"}]}]}`
	p, before := canonRoadmap(t, body)
	err := Edit(p, EditRequest{Slug: "checkout-ui", DependsOn: []string{"auth-service"}, SetDeps: true})
	if err == nil || !strings.Contains(err.Error(), "cycle detected") {
		t.Fatalf("multi-hop cycle must error, got %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, p)) {
		t.Fatal("rejected edit must be byte-identical")
	}
}
