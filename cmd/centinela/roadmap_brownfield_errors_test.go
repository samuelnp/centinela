package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func TestRoadmapBrownfield_MissingInventoryGuidesToAnalyze(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "draft.json")
	_, err := runBrown(t, filepath.Join(dir, "absent.json"), out, false, nil)
	if err == nil || !strings.Contains(err.Error(), "centinela analyze") {
		t.Fatalf("missing inventory must guide to analyze, got %v", err)
	}
	if _, statErr := os.Stat(out); statErr == nil {
		t.Fatal("no draft must be written on missing-inventory failure")
	}
}

func TestRoadmapBrownfield_MalformedInventoryErrors(t *testing.T) {
	in := writeInventory(t, "{bad")
	_, err := runBrown(t, in, filepath.Join(t.TempDir(), "draft.json"), false, nil)
	if err == nil || strings.Contains(err.Error(), "centinela analyze") {
		t.Fatalf("malformed inventory must be a distinct error, got %v", err)
	}
}

func TestRoadmapBrownfield_RefusesCanonicalOut(t *testing.T) {
	in := writeInventory(t, ntierInventory)
	// --out pointing at the canonical roadmap path must be refused, wrapped as a
	// write error, leaving no file at that path.
	out := roadmap.RoadmapFile
	_, err := runBrown(t, in, out, false, nil)
	if err == nil || !strings.Contains(err.Error(), "cannot write draft") {
		t.Fatalf("canonical --out must be refused as a write error, got %v", err)
	}
	if _, statErr := os.Stat(out); statErr == nil {
		t.Fatal("refused canonical --out must not create the file")
	}
}
