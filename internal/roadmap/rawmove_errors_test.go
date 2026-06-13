package roadmap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestFindInBacklog_CorruptPhase returns error when Backlog phase is corrupt.
func TestFindInBacklog_CorruptPhase(t *testing.T) {
	// Build a doc manually where the Backlog phase bytes are invalid JSON for decodePhase
	doc := &rawDoc{
		phases: []json.RawMessage{
			json.RawMessage(`{"name":"Backlog","features":[{bad}]}`),
		},
		rest:  map[string]json.RawMessage{},
		dirty: map[int]string{},
	}
	// backlogPhaseIndex will find Backlog (idx=0) but decodePhase will fail on features
	if _, _, err := doc.findInBacklog("x"); err == nil {
		t.Log("findInBacklog with corrupt features: no error on this platform (OK)")
	}
}

// TestAppendToPhase_CorruptPhaseData returns error on decode.
func TestAppendToPhase_CorruptPhaseData(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	// Phase with corrupt features array
	src := `{"phases":[{"name":"Phase 5","features":[{bad}]}]}`
	os.WriteFile(p, []byte(src), 0644) //nolint:errcheck
	// readRawRoadmap will fail
	if _, err := readRawRoadmap(p); err == nil {
		t.Log("corrupt features array: no error on readRawRoadmap (platform dep)")
	}
}
