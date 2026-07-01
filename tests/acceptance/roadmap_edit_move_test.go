package acceptance_test

// Acceptance: specs/roadmap-edit-move.feature
// Scenario: edit changes only the flags provided, leaving unspecified fields intact
// Scenario: edit --name renames the feature and rewrites dependents' dependsOn across ALL phases
// Scenario: edit --name refuses an invalid slug
// Scenario: edit --name refuses a collision with an existing feature, naming the owning phase
// Scenario: update is an alias for edit
// Scenario: edit/update a slug that does not exist errors "not found"

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

const emAccBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"auth-service"},` +
	`{"name":"checkout-ui","dependsOn":["auth-service"]}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"billing-api","dependsOn":["auth-service"]}]}]}`

// emPath returns the roadmap.json path inside a temp project dir.
func emPath(d string) string { return filepath.Join(d, ".workflow", "roadmap.json") }

// emFeat finds a feature by name in the on-disk roadmap of project d.
func emFeat(t *testing.T, d, name string) roadmap.Feature {
	t.Helper()
	b, _ := os.ReadFile(emPath(d))
	var r roadmap.Roadmap
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, p := range r.Phases {
		for _, f := range p.Features {
			if f.Name == name {
				return f
			}
		}
	}
	return roadmap.Feature{}
}

// TestAcc_EditRenameRewritesDependents drives rename through the binary.
func TestAcc_EditRenameRewritesDependents(t *testing.T) {
	d := rmcProject(t, emAccBody)
	if _, _, code := rmcRun(t, d, "roadmap", "edit", "auth-service", "--name", "auth-v2"); code != 0 {
		t.Fatalf("edit rename exit=%d", code)
	}
	if emFeat(t, d, "auth-service").Name != "" {
		t.Fatal("old name must be gone")
	}
	if d1 := emFeat(t, d, "checkout-ui").DependsOn; len(d1) != 1 || d1[0] != "auth-v2" {
		t.Fatalf("same-phase dependent not rewritten: %v", d1)
	}
	if d2 := emFeat(t, d, "billing-api").DependsOn; len(d2) != 1 || d2[0] != "auth-v2" {
		t.Fatalf("cross-phase dependent not rewritten: %v", d2)
	}
}

// TestAcc_EditAliasUpdate confirms `update` is an alias that edits a field.
func TestAcc_EditAliasUpdate(t *testing.T) {
	d := rmcProject(t, emAccBody)
	if _, _, code := rmcRun(t, d, "roadmap", "update", "auth-service", "--description", "Via alias"); code != 0 {
		t.Fatalf("update alias exit=%d", code)
	}
	if got := emFeat(t, d, "auth-service").Description; got != "Via alias" {
		t.Fatalf("alias must edit: %q", got)
	}
}

// TestAcc_EditRejectionsByteIdentical drives every edit refusal via the binary.
func TestAcc_EditRejectionsByteIdentical(t *testing.T) {
	rows := [][]string{
		{"roadmap", "edit", "auth-service", "--name", "Not_Kebab!"},
		{"roadmap", "edit", "auth-service", "--name", "billing-api"},
		{"roadmap", "edit", "ghost-feature", "--description", "x"},
		{"roadmap", "edit", "checkout-ui", "--depends-on", "checkout-ui"},
	}
	for _, args := range rows {
		d := rmcProject(t, emAccBody)
		before, _ := os.ReadFile(emPath(d))
		if _, _, code := rmcRun(t, d, args...); code == 0 {
			t.Fatalf("row %v must be rejected", args)
		}
		after, _ := os.ReadFile(emPath(d))
		if string(before) != string(after) {
			t.Fatalf("row %v must be byte-identical", args)
		}
	}
}
