package roadmap

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

// writeArtifact must surface an indentValue error for a malformed non-features key.
func TestWriteArtifact_EmitIndentError(t *testing.T) {
	p := filepath.Join(t.TempDir(), "art.json")
	top := map[string]json.RawMessage{"role": json.RawMessage("{bad}")}
	if err := writeArtifact(p, top); err == nil {
		t.Fatal("expected indentValue error for malformed non-features key")
	}
}

// writeArtifact must surface a writeFeatureArray error when features is not an array.
func TestWriteArtifact_FeaturesNotArray(t *testing.T) {
	p := filepath.Join(t.TempDir(), "art.json")
	top := map[string]json.RawMessage{"features": json.RawMessage("42")}
	if err := writeArtifact(p, top); err == nil {
		t.Fatal("expected writeFeatureArray error for non-array features")
	}
}

// appendFeatureEntry must surface a compactBytes error when the entry is invalid.
func TestAppendFeatureEntry_CompactError(t *testing.T) {
	p := filepath.Join(t.TempDir(), "art.json")
	if err := writeArtifact(p, map[string]json.RawMessage{
		"features": json.RawMessage("[]"),
	}); err != nil {
		t.Fatalf("seed writeArtifact: %v", err)
	}
	if err := appendFeatureEntry(p, json.RawMessage("{bad}")); err == nil {
		t.Fatal("expected compactBytes error for an invalid appended entry")
	}
}
