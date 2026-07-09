package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

const phaseAddBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"a"}]},` +
	`{"name":"Backlog","features":[]}]}`

// readBytes returns the on-disk bytes at path, failing the test on error.
func readBytes(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}

// resetPhaseAddFlags restores the add command globals after a test.
func resetPhaseAddFlags() { phaseAddNote, phaseAddAfter = "", "" }

// TestRunRoadmapPhaseAdd_WithFlags inserts via --after and --note.
func TestRunRoadmapPhaseAdd_WithFlags(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, phaseAddBody)
	defer resetPhaseAddFlags()
	phaseAddAfter, phaseAddNote = "Phase 1: Foundations", "rationale"
	if err := runRoadmapPhaseAdd(nil, []string{"Phase 1.5: Bridge"}); err != nil {
		t.Fatalf("runRoadmapPhaseAdd: %v", err)
	}
	got := string(readBytes(t, roadmap.RoadmapFile))
	if !strings.Contains(got, "Phase 1.5: Bridge") || !strings.Contains(got, `"note": "rationale"`) {
		t.Fatalf("phase/note missing: %s", got)
	}
}

// TestRunRoadmapPhaseAdd_Error surfaces a package rejection (duplicate name).
func TestRunRoadmapPhaseAdd_Error(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, phaseAddBody)
	defer resetPhaseAddFlags()
	if err := runRoadmapPhaseAdd(nil, []string{"Phase 1: Foundations"}); err == nil {
		t.Fatal("duplicate phase name must error")
	}
}
