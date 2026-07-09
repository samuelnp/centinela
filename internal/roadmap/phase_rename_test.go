package roadmap

import (
	"bytes"
	"strings"
	"testing"
)

// phRenameBody: Phase 1 carries two features (to prove they survive a rename), plus
// an untouched Phase 2 and a Backlog.
const phRenameBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"auth-service"},{"name":"checkout-ui"}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"b"}]},` +
	`{"name":"Backlog","features":[]}]}`

// TestPhaseRename_InPlace renames the phase, keeps its features, and leaves other
// phases byte-identical.
func TestPhaseRename_InPlace(t *testing.T) {
	p, before := canonRoadmap(t, phRenameBody)
	if err := PhaseRename(p, "Phase 1: Foundations", "Phase 1: Core"); err != nil {
		t.Fatalf("PhaseRename: %v", err)
	}
	names := phaseOrderNames(t, p)
	if strings.Join(names, ",") != "Phase 1: Core,Phase 2: Growth,Backlog" {
		t.Fatalf("order wrong: %v", names)
	}
	feats := orderIn(t, p, "Phase 1: Core")
	if strings.Join(feats, ",") != "auth-service,checkout-ui" {
		t.Fatalf("features must survive rename: %v", feats)
	}
	if !bytes.Contains(crudBytes(t, p), phaseSlice(t, before, "Phase 2: Growth")) {
		t.Fatal("Phase 2 must be byte-identical")
	}
}

// TestPhaseRename_SameNameNoop: same name → byte-identical, no write.
func TestPhaseRename_SameNameNoop(t *testing.T) {
	p, before := canonRoadmap(t, phRenameBody)
	if err := PhaseRename(p, "Phase 1: Foundations", "Phase 1: Foundations"); err != nil {
		t.Fatalf("same-name rename must be a no-op: %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, p)) {
		t.Fatal("same-name rename must be byte-identical")
	}
}

// TestPhaseRename_Refusals: unknown old, collision, empty new, either side reserved,
// each byte-identical.
func TestPhaseRename_Refusals(t *testing.T) {
	cases := []struct{ old, name, sub string }{
		{"Phase 9: Nope", "Phase 3: Scale", "not found"},
		{"Phase 1: Foundations", "Phase 2: Growth", "already exists"},
		{"Phase 1: Foundations", "", "phase name is required"},
		{"Backlog", "Phase 3: Scale", "reserved phase name"},
		{"Phase 1: Foundations", "Baseline", "reserved phase name"},
	}
	for _, c := range cases {
		p, before := canonRoadmap(t, phRenameBody)
		wantErr(t, PhaseRename(p, c.old, c.name), c.sub)
		if !bytes.Equal(before, crudBytes(t, p)) {
			t.Fatalf("refused rename %q->%q must be byte-identical", c.old, c.name)
		}
	}
}
