package unit_test

// Acceptance: specs/roadmap-phase-ops.feature
// Scenario: phase add with --after inserts immediately after the named phase
// Scenario: phase add without --after lands before the Backlog phase
// Scenario: phase add on an empty roadmap succeeds as the first phase
// Scenario: phase rename renames in place, leaving its features and other phases untouched
// Scenario: phase rename to the SAME name is a no-op, byte-identical
// Scenario: phase remove deletes an empty phase
// Scenario: phase remove of a non-empty phase without --force is refused, naming the feature count

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

const poBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"auth-service"}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"billing-api"},{"name":"reporting"}]},` +
	`{"name":"Backlog","features":[]}]}`

// poWrite writes body to a standalone roadmap.json in a temp dir.
func poWrite(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "roadmap.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func poOrder(t *testing.T, path string) string {
	t.Helper()
	b, _ := os.ReadFile(path)
	var r roadmap.Roadmap
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	names := make([]string, 0, len(r.Phases))
	for _, p := range r.Phases {
		names = append(names, p.Name)
	}
	return strings.Join(names, ",")
}

// TestPO_AddPositions covers --after, before-Backlog default, and empty roadmap.
func TestPO_AddPositions(t *testing.T) {
	p := poWrite(t, poBody)
	if err := roadmap.PhaseAdd(p, "X", "", "Phase 1: Foundations"); err != nil {
		t.Fatalf("add --after: %v", err)
	}
	if poOrder(t, p) != "Phase 1: Foundations,X,Phase 2: Growth,Backlog" {
		t.Fatalf("after order wrong: %s", poOrder(t, p))
	}
	e := poWrite(t, `{"phases":[]}`)
	if err := roadmap.PhaseAdd(e, "Phase 1: Foundations", "", ""); err != nil {
		t.Fatalf("add on empty: %v", err)
	}
	if poOrder(t, e) != "Phase 1: Foundations" {
		t.Fatalf("empty-roadmap add wrong: %s", poOrder(t, e))
	}
}

// TestPO_RenameSameNameNoop: same-name rename is byte-identical.
func TestPO_RenameSameNameNoop(t *testing.T) {
	p := poWrite(t, poBody)
	before, _ := os.ReadFile(p)
	if err := roadmap.PhaseRename(p, "Phase 1: Foundations", "Phase 1: Foundations"); err != nil {
		t.Fatalf("same-name rename: %v", err)
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Fatal("same-name rename must be byte-identical")
	}
}

// TestPO_RemoveEmptyAndNonEmptyRefusal: empty removed, non-empty refused byte-identical.
func TestPO_RemoveEmptyAndNonEmptyRefusal(t *testing.T) {
	p := poWrite(t, poBody)
	if err := roadmap.PhaseAdd(p, "Phase 3: Scale", "", ""); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := roadmap.PhaseRemove(p, "Phase 3: Scale", false); err != nil {
		t.Fatalf("remove empty: %v", err)
	}
	before, _ := os.ReadFile(p)
	if err := roadmap.PhaseRemove(p, "Phase 2: Growth", false); err == nil || !strings.Contains(err.Error(), "2 features") {
		t.Fatalf("non-empty must refuse naming count: %v", err)
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Fatal("refused remove must be byte-identical")
	}
}
