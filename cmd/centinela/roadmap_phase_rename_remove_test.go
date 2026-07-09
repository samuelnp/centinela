package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// TestRunRoadmapPhaseRename_Success renames a phase through the command.
func TestRunRoadmapPhaseRename_Success(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, phaseAddBody)
	if err := runRoadmapPhaseRename(nil, []string{"Phase 1: Foundations", "Phase 1: Core"}); err != nil {
		t.Fatalf("runRoadmapPhaseRename: %v", err)
	}
	if got := string(readBytes(t, roadmap.RoadmapFile)); !strings.Contains(got, "Phase 1: Core") {
		t.Fatalf("rename not applied: %s", got)
	}
}

// TestRunRoadmapPhaseRename_Error surfaces a package rejection (unknown old name).
func TestRunRoadmapPhaseRename_Error(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, phaseAddBody)
	if err := runRoadmapPhaseRename(nil, []string{"Phase 9: Nope", "Phase 3: Scale"}); err == nil {
		t.Fatal("unknown old phase must error")
	}
}

// resetPhaseRemoveFlags restores the remove command global after a test.
func resetPhaseRemoveFlags() { phaseRemoveForce = false }

// TestRunRoadmapPhaseRemove_Success removes an empty phase through the command.
func TestRunRoadmapPhaseRemove_Success(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, `{"phases":[`+
		`{"name":"Phase 1: Foundations","features":[]},{"name":"Backlog","features":[]}]}`)
	defer resetPhaseRemoveFlags()
	if err := runRoadmapPhaseRemove(nil, []string{"Phase 1: Foundations"}); err != nil {
		t.Fatalf("runRoadmapPhaseRemove: %v", err)
	}
	if got := string(readBytes(t, roadmap.RoadmapFile)); strings.Contains(got, "Phase 1: Foundations") {
		t.Fatalf("phase must be gone: %s", got)
	}
}

// TestRunRoadmapPhaseRemove_ForceFlag: without --force a non-empty phase is refused;
// with --force (draft feature, so no analysis/quality coverage needed) it removes.
func TestRunRoadmapPhaseRemove_ForceFlag(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, `{"phases":[`+
		`{"name":"Phase 1: Foundations","features":[{"name":"a","draft":true}]},`+
		`{"name":"Backlog","features":[]}]}`)
	writeFile(t, roadmap.RoadmapAnalysisFile, `{"role":"senior-product-manager","features":[]}`)
	writeFile(t, roadmap.RoadmapQualityFile, `{"role":"roadmap-quality-evaluator","threshold":9,"features":[]}`)
	writeFile(t, roadmap.RoadmapAnalysisMarkdown, "# a\n")
	writeFile(t, roadmap.RoadmapQualityMarkdown, "# q\n")
	defer resetPhaseRemoveFlags()
	if err := runRoadmapPhaseRemove(nil, []string{"Phase 1: Foundations"}); err == nil {
		t.Fatal("non-empty phase without --force must error")
	}
	phaseRemoveForce = true
	if err := runRoadmapPhaseRemove(nil, []string{"Phase 1: Foundations"}); err != nil {
		t.Fatalf("--force remove: %v", err)
	}
	if got := string(readBytes(t, roadmap.RoadmapFile)); strings.Contains(got, "Phase 1: Foundations") {
		t.Fatalf("phase must be gone: %s", got)
	}
}
