package roadmap

import (
	"encoding/json"
	"testing"
)

// newRawDoc builds a rawDoc from literal phase JSON strings for white-box tests.
func newRawDoc(raw ...string) *rawDoc {
	ph := make([]json.RawMessage, len(raw))
	for i, r := range raw {
		ph[i] = json.RawMessage(r)
	}
	return &rawDoc{phases: ph, rest: map[string]json.RawMessage{}, dirty: map[int]string{}}
}

// phaseFeatureNames must surface a decodePhase error when a phase is malformed.
func TestPhaseFeatureNames_DecodeError(t *testing.T) {
	doc := newRawDoc(`{"name":"P","features":42}`)
	if _, err := doc.phaseFeatureNames(); err == nil {
		t.Fatal("expected decodePhase error for non-array features")
	}
}

// phaseFeatureNames must surface a featureName error for a non-object entry.
func TestPhaseFeatureNames_FeatureNameError(t *testing.T) {
	doc := newRawDoc(`{"name":"P","features":[42]}`)
	if _, err := doc.phaseFeatureNames(); err == nil {
		t.Fatal("expected featureName error for numeric feature entry")
	}
}

// setPhase must surface an encodePhase error for an invalid raw feature.
func TestSetPhase_EncodeError(t *testing.T) {
	doc := newRawDoc()
	p := &rawPhase{Name: "P", Features: []json.RawMessage{json.RawMessage("{bad}")}}
	if err := doc.setPhase(0, p); err == nil {
		t.Fatal("expected encodePhase error for invalid feature bytes")
	}
}

// appendBacklog creating a fresh Backlog phase must surface an encode error.
func TestAppendBacklog_NewPhaseEncodeError(t *testing.T) {
	doc := newRawDoc() // no phases -> backlog absent -> new-phase branch
	if err := doc.appendBacklog(json.RawMessage("{bad}")); err == nil {
		t.Fatal("expected encode error when building a new Backlog phase")
	}
}

// appendBacklog into an existing malformed Backlog phase must surface a decode error.
func TestAppendBacklog_ExistingPhaseDecodeError(t *testing.T) {
	doc := newRawDoc(`{"name":"Backlog","features":42}`)
	entry, _ := compactBytes(Feature{Name: "x"})
	if err := doc.appendBacklog(entry); err == nil {
		t.Fatal("expected decodePhase error for malformed Backlog phase")
	}
}

// findInBacklog must surface a decodePhase error for a malformed Backlog phase.
func TestFindInBacklog_DecodeError(t *testing.T) {
	doc := newRawDoc(`{"name":"Backlog","features":42}`)
	if _, _, err := doc.findInBacklog("x"); err == nil {
		t.Fatal("expected decodePhase error from findInBacklog")
	}
}

// removeBacklogFeature must surface a decodePhase error for a malformed phase.
func TestRemoveBacklogFeature_DecodeError(t *testing.T) {
	doc := newRawDoc(`{"name":"Backlog","features":42}`)
	if err := doc.removeBacklogFeature(0, "x"); err == nil {
		t.Fatal("expected decodePhase error from removeBacklogFeature")
	}
}
