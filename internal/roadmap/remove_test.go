package roadmap

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestRemove_Success deletes a planned, undepended feature and keeps the file valid.
func TestRemove_Success(t *testing.T) {
	// lonely-feature in Phase 3 has no dependents and is planned.
	p := crudWrite(t, crudBody)
	if err := Remove(p, "lonely-feature"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	data := crudBytes(t, p)
	if bytes.Contains(data, []byte("lonely-feature")) {
		t.Fatal("removed feature must be gone")
	}
	var r Roadmap
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatalf("result must be valid JSON: %v", err)
	}
	// Untouched phases survive byte-stably.
	if !strings.Contains(string(data), "billing-api") {
		t.Fatal("untouched phase feature must survive remove")
	}
}

// TestRemove_LastFeatureLeavesEmptyArray keeps the now-empty phase present.
func TestRemove_LastFeatureLeavesEmptyArray(t *testing.T) {
	p := crudWrite(t, crudBody)
	if err := Remove(p, "lonely-feature"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	s := string(crudBytes(t, p))
	if !strings.Contains(s, "Phase 3: Solo") {
		t.Fatal("empty phase must remain present")
	}
	if !strings.Contains(s, `"features": []`) {
		t.Fatalf("phase must keep an empty features array: %s", s)
	}
}

// TestRemove_NotFound errors and leaves roadmap.json byte-identical.
func TestRemove_NotFound(t *testing.T) {
	p := crudWrite(t, crudBody)
	before := crudBytes(t, p)
	err := Remove(p, "ghost-feature")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, p)) {
		t.Fatal("failed remove must leave roadmap.json byte-identical")
	}
}
