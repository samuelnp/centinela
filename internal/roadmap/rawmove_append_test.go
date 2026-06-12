package roadmap

import (
	"os"
	"strings"
	"testing"
)

// TestAppendToPhase_HappyPath moves slug into target phase.
func TestAppendToPhase_HappyPath(t *testing.T) {
	_, p := roadmapWithBacklog(t, "my-finding")
	doc, _ := readRawRoadmap(p)
	if err := doc.appendToPhase("Phase 5", "my-finding"); err != nil {
		t.Fatalf("appendToPhase: %v", err)
	}
	writeRawRoadmap(p, doc) //nolint:errcheck
	data, _ := os.ReadFile(p)
	if !strings.Contains(string(data), `"my-finding"`) {
		t.Errorf("slug must appear in Phase 5: %s", data)
	}
}

// TestAppendToPhase_UnknownPhase returns error with known phase names.
func TestAppendToPhase_UnknownPhase(t *testing.T) {
	_, p := roadmapWithBacklog(t, "my-finding")
	doc, _ := readRawRoadmap(p)
	err := doc.appendToPhase("Phase 99", "my-finding")
	if err == nil {
		t.Fatal("expected error for unknown phase")
	}
	if !strings.Contains(err.Error(), "Phase 5") {
		t.Errorf("error must list known phases, got: %v", err)
	}
}

// TestAppendToPhase_RefusesToAddToBacklog returns error when target is Backlog.
func TestAppendToPhase_RefusesToAddToBacklog(t *testing.T) {
	_, p := roadmapWithBacklog(t, "x")
	doc, _ := readRawRoadmap(p)
	err := doc.appendToPhase("Backlog", "x")
	if err == nil {
		t.Error("appending to Backlog must be refused (Backlog is not a valid target)")
	}
}

// TestAppendToPhase_DuplicateInTarget refuses duplicate slug in target phase.
func TestAppendToPhase_DuplicateInTarget(t *testing.T) {
	_, p := roadmapWithBacklog(t, "my-finding")
	doc, _ := readRawRoadmap(p)
	err := doc.appendToPhase("Phase 5", "existing")
	if err == nil {
		t.Error("duplicate slug in target phase must be refused")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error must mention duplicate, got: %v", err)
	}
}
