package roadmap

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// renameBody: a dependent in each phase so a rename must rewrite dependsOn
// across ALL phases.
const renameBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"auth-service"},` +
	`{"name":"checkout-ui","dependsOn":["auth-service"]}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"billing-api","dependsOn":["auth-service"]}]}]}`

// TestEdit_RenameRewritesDependentsAcrossPhases renames and rewrites every
// dependent across phases, leaving the roadmap dependency-valid.
func TestEdit_RenameRewritesDependentsAcrossPhases(t *testing.T) {
	p, _ := canonRoadmap(t, renameBody)
	if err := Edit(p, EditRequest{Slug: "auth-service", NewName: "auth-service-v2"}); err != nil {
		t.Fatalf("Edit rename: %v", err)
	}
	ph1 := orderIn(t, p, "Phase 1: Foundations")
	if !contains(ph1, "auth-service-v2") || contains(ph1, "auth-service") {
		t.Fatalf("feature must be renamed: %v", ph1)
	}
	if d := featureIn(t, p, "checkout-ui").DependsOn; len(d) != 1 || d[0] != "auth-service-v2" {
		t.Fatalf("same-phase dependent not rewritten: %v", d)
	}
	if d := featureIn(t, p, "billing-api").DependsOn; len(d) != 1 || d[0] != "auth-service-v2" {
		t.Fatalf("cross-phase dependent not rewritten: %v", d)
	}
	var r Roadmap
	if err := json.Unmarshal(crudBytes(t, p), &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := ValidateDependencies(&r); err != nil {
		t.Fatalf("validate must PASS after rename: %v", err)
	}
}

// TestEdit_RenameCollisionByteIdentical refuses a rename onto an existing name,
// naming the owning phase, and writes nothing.
func TestEdit_RenameCollisionByteIdentical(t *testing.T) {
	p, before := canonRoadmap(t, renameBody)
	err := Edit(p, EditRequest{Slug: "auth-service", NewName: "billing-api"})
	if err == nil || !strings.Contains(err.Error(), `"billing-api" already exists in phase "Phase 2: Growth"`) {
		t.Fatalf("collision must name the owning phase, got %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, p)) {
		t.Fatal("rejected rename must be byte-identical")
	}
}

// TestEdit_RenameSameNameStable treats an unchanged name as a semantic no-op: no
// dependents are rewritten and the write is idempotent. Note Edit always re-renders
// the target's phase one-per-line, so the result is NOT byte-identical against a
// json.Indent-canonical file (unlike a rejected op) — see qa-senior deferred
// findings for the spec's "same-name byte-identical" scenario.
func TestEdit_RenameSameNameStable(t *testing.T) {
	p, _ := canonRoadmap(t, renameBody)
	if err := Edit(p, EditRequest{Slug: "auth-service", NewName: "auth-service"}); err != nil {
		t.Fatalf("Edit same-name: %v", err)
	}
	if d := featureIn(t, p, "checkout-ui").DependsOn; len(d) != 1 || d[0] != "auth-service" {
		t.Fatalf("same-name must not rewrite same-phase dependent: %v", d)
	}
	if d := featureIn(t, p, "billing-api").DependsOn; len(d) != 1 || d[0] != "auth-service" {
		t.Fatalf("same-name must not rewrite cross-phase dependent: %v", d)
	}
	settled := crudBytes(t, p)
	if err := Edit(p, EditRequest{Slug: "auth-service", NewName: "auth-service"}); err != nil {
		t.Fatalf("Edit same-name (2nd): %v", err)
	}
	if !bytes.Equal(settled, crudBytes(t, p)) {
		t.Fatal("same-name edit must be idempotent once settled")
	}
}

// TestEdit_RenameInvalidSlugByteIdentical refuses a non-kebab slug, writing nothing.
func TestEdit_RenameInvalidSlugByteIdentical(t *testing.T) {
	p, before := canonRoadmap(t, renameBody)
	err := Edit(p, EditRequest{Slug: "auth-service", NewName: "Not_Kebab!"})
	if err == nil || !strings.Contains(err.Error(), "invalid feature slug") {
		t.Fatalf("invalid slug must error, got %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, p)) {
		t.Fatal("rejected rename must be byte-identical")
	}
}
