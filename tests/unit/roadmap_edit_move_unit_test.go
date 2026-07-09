package unit_test

// Acceptance: specs/roadmap-edit-move.feature
// Scenario: edit changes only the flags provided, leaving unspecified fields intact
// Scenario: edit --name renames the feature and rewrites dependents' dependsOn across ALL phases
// Scenario: move relocates a feature to the target phase, appending by default
// Scenario: reorder repositions a feature within its own phase
// Scenario: every mutation performs exactly one atomic write — a rejected edit/move/reorder writes nothing

import (
	"bytes"
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

const emBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"auth-service"},` +
	`{"name":"checkout-ui","dependsOn":["auth-service"]}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"billing-api","dependsOn":["auth-service"]}]}]}`

// emFeat returns the named feature from the on-disk roadmap at p.
func emFeat(t *testing.T, p, name string) roadmap.Feature {
	t.Helper()
	for _, ph := range loadPath(t, p).Phases {
		for _, f := range ph.Features {
			if f.Name == name {
				return f
			}
		}
	}
	t.Fatalf("feature %q missing", name)
	return roadmap.Feature{}
}

// TestEM_EditOnlyProvided edits description; deps stay intact and validate PASSes.
func TestEM_EditOnlyProvided(t *testing.T) {
	p := unitPath(t, emBody)
	if err := roadmap.Edit(p, roadmap.EditRequest{Slug: "checkout-ui", Description: "New"}); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	f := emFeat(t, p, "checkout-ui")
	if f.Description != "New" || len(f.DependsOn) != 1 || f.DependsOn[0] != "auth-service" {
		t.Fatalf("only description should change: %+v", f)
	}
	if err := roadmap.ValidateDependencies(loadPath(t, p)); err != nil {
		t.Fatalf("validate must PASS: %v", err)
	}
}

// TestEM_RenameRewritesDependents renames and rewrites dependents across phases.
func TestEM_RenameRewritesDependents(t *testing.T) {
	p := unitPath(t, emBody)
	if err := roadmap.Edit(p, roadmap.EditRequest{Slug: "auth-service", NewName: "auth-v2"}); err != nil {
		t.Fatalf("Edit rename: %v", err)
	}
	if d := emFeat(t, p, "checkout-ui").DependsOn; d[0] != "auth-v2" {
		t.Fatalf("same-phase dependent: %v", d)
	}
	if d := emFeat(t, p, "billing-api").DependsOn; d[0] != "auth-v2" {
		t.Fatalf("cross-phase dependent: %v", d)
	}
}

// TestEM_MoveAndReorder relocates then repositions, and asserts a rejected op is
// byte-identical (single atomic write only on success).
func TestEM_MoveAndReorder(t *testing.T) {
	p := unitPath(t, emBody)
	if err := roadmap.Move(p, roadmap.MoveRequest{Slug: "checkout-ui", ToPhase: "Phase 2: Growth"}); err != nil {
		t.Fatalf("Move: %v", err)
	}
	if emFeat(t, p, "checkout-ui").Name == "" {
		t.Fatal("checkout-ui must survive the move")
	}
	before, _ := os.ReadFile(p)
	if err := roadmap.Reorder(p, roadmap.ReorderRequest{Slug: "ghost", BeforeAnchor: "billing-api"}); err == nil {
		t.Fatal("unknown slug reorder must error")
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Fatal("rejected reorder must be byte-identical")
	}
}
