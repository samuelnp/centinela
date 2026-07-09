package roadmap

import (
	"encoding/json"
	"testing"
)

// TestPhaseOps_MalformedRoadmapErrors: every op propagates a read/parse error from
// a malformed roadmap.json, exercising the readRawRoadmap error branches.
func TestPhaseOps_MalformedRoadmapErrors(t *testing.T) {
	p := crudWrite(t, "{ not valid json")
	if err := PhaseAdd(p, "Phase 9: X", "", ""); err == nil {
		t.Fatal("PhaseAdd on malformed roadmap must error")
	}
	if err := PhaseRename(p, "Phase 1: Foundations", "Phase 9: X"); err == nil {
		t.Fatal("PhaseRename on malformed roadmap must error")
	}
	if err := PhaseRemove(p, "Phase 1: Foundations", false); err == nil {
		t.Fatal("PhaseRemove on malformed roadmap must error")
	}
}

// TestPhaseOps_MissingRoadmapErrors: a missing file surfaces an error, no panic.
func TestPhaseOps_MissingRoadmapErrors(t *testing.T) {
	if err := PhaseAdd(t.TempDir()+"/nope.json", "Phase 9: X", "", ""); err == nil {
		t.Fatal("PhaseAdd on missing roadmap must error")
	}
}

// TestReindex_DirtyBeforeInsertUnshifted: a dirty phase at an index BELOW the insert
// position keeps its key (the k < pos branch of insertPhaseAt), and a later insert
// keeps the earlier dirty render intact.
func TestReindex_DirtyBeforeInsertUnshifted(t *testing.T) {
	doc, err := readRawRoadmap(crudWrite(t, reindexBody))
	if err != nil {
		t.Fatalf("readRawRoadmap: %v", err)
	}
	dirtyLater(t, doc, 0) // dirty "Phase 1: Foundations" at index 0
	// insert AFTER the dirty phase -> its key stays 0 (unshifted branch)
	if err := doc.insertPhaseAt(2, json.RawMessage(`{"name":"Mid","features":[]}`)); err != nil {
		t.Fatalf("insertPhaseAt: %v", err)
	}
	if _, ok := doc.dirty[0]; !ok {
		t.Fatal("dirty key 0 must survive an insert after it")
	}
	if _, err := doc.render(); err != nil {
		t.Fatalf("render: %v", err)
	}
}

// TestReindex_DirtyBeforeRemoveUnshifted: a dirty phase below the removed index keeps
// its key (the default branch of removePhaseAt).
func TestReindex_DirtyBeforeRemoveUnshifted(t *testing.T) {
	doc, err := readRawRoadmap(crudWrite(t, reindexBody))
	if err != nil {
		t.Fatalf("readRawRoadmap: %v", err)
	}
	dirtyLater(t, doc, 0)                        // dirty index 0
	if err := doc.removePhaseAt(2); err != nil { // remove a later phase
		t.Fatalf("removePhaseAt: %v", err)
	}
	if _, ok := doc.dirty[0]; !ok {
		t.Fatal("dirty key 0 must survive removal of a later phase")
	}
}

// TestRenamePhaseAt_DecodeError: renaming an index holding a malformed phase entry
// surfaces the decode error branch.
func TestRenamePhaseAt_DecodeError(t *testing.T) {
	doc, err := readRawRoadmap(crudWrite(t, `{"phases":[{"name":"Phase 1: Foundations","features":[]}]}`))
	if err != nil {
		t.Fatalf("readRawRoadmap: %v", err)
	}
	doc.phases[0] = json.RawMessage(`{"name":123}`) // malformed: name not a string
	if err := doc.renamePhaseAt(0, "X"); err == nil {
		t.Fatal("renamePhaseAt on a malformed phase must error")
	}
}

// TestPhaseIndexByName_MalformedError: a malformed phase entry surfaces the
// phaseName error branch of phaseIndexByName.
func TestPhaseIndexByName_MalformedError(t *testing.T) {
	doc, err := readRawRoadmap(crudWrite(t, `{"phases":[{"name":"Phase 1: Foundations","features":[]}]}`))
	if err != nil {
		t.Fatalf("readRawRoadmap: %v", err)
	}
	doc.phases[0] = json.RawMessage(`{"name":123}`)
	if _, err := doc.phaseIndexByName("anything"); err == nil {
		t.Fatal("phaseIndexByName over a malformed phase must error")
	}
}
