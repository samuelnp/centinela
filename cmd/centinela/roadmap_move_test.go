package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// resetMoveFlags restores the move command's globals after a test.
func resetMoveFlags() { moveToPhase, moveBefore, moveAfter = "", "", "" }

// TestRunRoadmapMove_Success relocates a feature to another phase.
func TestRunRoadmapMove_Success(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, addRoadmap)
	defer resetMoveFlags()
	moveToPhase = "Phase 2"
	if err := runRoadmapMove(nil, []string{"auth-service"}); err != nil {
		t.Fatalf("runRoadmapMove: %v", err)
	}
	for _, p := range mustLoad(t).Phases {
		if p.Name == "Phase 1" {
			for _, f := range p.Features {
				if f.Name == "auth-service" {
					t.Fatal("auth-service must leave Phase 1")
				}
			}
		}
	}
}

// TestRunRoadmapMove_PhaseRequired errors when --to-phase is unset.
func TestRunRoadmapMove_PhaseRequired(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, addRoadmap)
	defer resetMoveFlags()
	moveToPhase = ""
	if err := runRoadmapMove(nil, []string{"auth-service"}); err == nil {
		t.Fatal("missing --to-phase must error")
	}
}

// TestRunRoadmapMove_MutuallyExclusive rejects --before with --after.
func TestRunRoadmapMove_MutuallyExclusive(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, addRoadmap)
	defer resetMoveFlags()
	moveToPhase, moveBefore, moveAfter = "Phase 2", "billing-api", "billing-api"
	if err := runRoadmapMove(nil, []string{"auth-service"}); err == nil {
		t.Fatal("--before and --after together must error")
	}
}

// TestRunRoadmapMove_Error surfaces a package rejection (unknown slug).
func TestRunRoadmapMove_Error(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, addRoadmap)
	defer resetMoveFlags()
	moveToPhase = "Phase 2"
	if err := runRoadmapMove(nil, []string{"ghost"}); err == nil {
		t.Fatal("unknown slug must error")
	}
}
