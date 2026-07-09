package roadmap

import (
	"encoding/json"
	"testing"
)

// reindexBody: three phases, an untouched middle "Phase 2: Growth" and a later
// "Phase 3: Scale" carrying "reporting" which the tests concurrently dirty.
const reindexBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"a"}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"mid"}]},` +
	`{"name":"Phase 3: Scale","features":[{"name":"reporting"}]}]}`

// dirtyLater mutates feature "reporting" in the phase currently at idx (marking it
// dirty), simulating a later-phase edit happening in the SAME run as a structural op.
func dirtyLater(t *testing.T, d *rawDoc, idx int) {
	t.Helper()
	p, err := d.decodePhase(idx)
	if err != nil {
		t.Fatalf("decodePhase: %v", err)
	}
	p.Features[0] = json.RawMessage(`{"name":"reporting","note":"edited"}`)
	if err := d.setPhase(idx, p); err != nil {
		t.Fatalf("setPhase: %v", err)
	}
}

// TestReindex_InsertEarlierWhileLaterDirty inserts a new first phase AFTER dirtying
// the last phase, then asserts the EXACT rendered bytes: the dirty phase must render
// at its shifted index, not corrupt "Phase 2: Growth". This is the reindex proof.
func TestReindex_InsertEarlierWhileLaterDirty(t *testing.T) {
	p := crudWrite(t, reindexBody)
	doc, err := readRawRoadmap(p)
	if err != nil {
		t.Fatalf("readRawRoadmap: %v", err)
	}
	dirtyLater(t, doc, 2) // dirty "Phase 3: Scale" (index 2)
	newPhase := json.RawMessage(`{"name":"Phase 0: Bootstrap","features":[]}`)
	if err := doc.insertPhaseAt(0, newPhase); err != nil {
		t.Fatalf("insertPhaseAt: %v", err)
	}
	got, err := doc.render()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := "{\n  \"phases\": [\n" +
		"    {\n      \"name\": \"Phase 0: Bootstrap\",\n      \"features\": []\n    },\n" +
		"    {\n      \"name\": \"Phase 1: Foundations\",\n      \"features\": [\n        {\n          \"name\": \"a\"\n        }\n      ]\n    },\n" +
		"    {\n      \"name\": \"Phase 2: Growth\",\n      \"features\": [\n        {\n          \"name\": \"mid\"\n        }\n      ]\n    },\n" +
		"    {\n      \"name\": \"Phase 3: Scale\",\n      \"features\": [\n        {\"name\":\"reporting\",\"note\":\"edited\"}\n      ]\n    }\n" +
		"  ]\n}\n"
	if string(got) != want {
		t.Fatalf("reindexed insert render wrong:\n got=%q\nwant=%q", got, want)
	}
}

// TestReindex_RemoveMiddleWhileLaterDirty removes the empty middle phase AFTER
// dirtying the last phase; the dirty phase must render at its shifted-down index.
func TestReindex_RemoveMiddleWhileLaterDirty(t *testing.T) {
	body := `{"phases":[` +
		`{"name":"Phase 1: Foundations","features":[{"name":"a"}]},` +
		`{"name":"Phase 2: Growth","features":[]},` +
		`{"name":"Phase 3: Scale","features":[{"name":"reporting"}]}]}`
	p := crudWrite(t, body)
	doc, err := readRawRoadmap(p)
	if err != nil {
		t.Fatalf("readRawRoadmap: %v", err)
	}
	dirtyLater(t, doc, 2)
	if err := doc.removePhaseAt(1); err != nil {
		t.Fatalf("removePhaseAt: %v", err)
	}
	got, err := doc.render()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := "{\n  \"phases\": [\n" +
		"    {\n      \"name\": \"Phase 1: Foundations\",\n      \"features\": [\n        {\n          \"name\": \"a\"\n        }\n      ]\n    },\n" +
		"    {\n      \"name\": \"Phase 3: Scale\",\n      \"features\": [\n        {\"name\":\"reporting\",\"note\":\"edited\"}\n      ]\n    }\n" +
		"  ]\n}\n"
	if string(got) != want {
		t.Fatalf("reindexed remove render wrong:\n got=%q\nwant=%q", got, want)
	}
}

// TestReindex_BoundsErrors: out-of-range positions/indices return errors, not panics.
func TestReindex_BoundsErrors(t *testing.T) {
	doc, err := readRawRoadmap(crudWrite(t, reindexBody))
	if err != nil {
		t.Fatalf("readRawRoadmap: %v", err)
	}
	wantErr(t, doc.insertPhaseAt(-1, json.RawMessage(`{}`)), "out of range")
	wantErr(t, doc.insertPhaseAt(99, json.RawMessage(`{}`)), "out of range")
	wantErr(t, doc.removePhaseAt(-1), "out of range")
	wantErr(t, doc.removePhaseAt(99), "out of range")
}
