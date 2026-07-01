package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

const addRoadmap = `{"phases":[{"name":"Phase 1","features":[{"name":"auth-service"}]},` +
	`{"name":"Phase 2","features":[{"name":"billing-api"}]}]}`

// resetAddFlags restores the add command's global flag vars after a test.
func resetAddFlags() {
	addPhase, addDescription, addArchetype, addDependsOn = "", "", "", nil
}

// TestRunRoadmapAdd_Success authors a draft and persists it.
func TestRunRoadmapAdd_Success(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, addRoadmap)
	defer resetAddFlags()
	addPhase = "Phase 1"
	addDescription = "Adds it"
	addDependsOn = []string{"auth-service"}
	if err := runRoadmapAdd(nil, []string{"new-widget"}); err != nil {
		t.Fatalf("runRoadmapAdd: %v", err)
	}
	r, err := roadmap.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !roadmap.IsDraftFeature(r, "new-widget") {
		t.Fatal("added feature must be a draft")
	}
}

// TestRunRoadmapAdd_PhaseRequired errors when --phase is unset.
func TestRunRoadmapAdd_PhaseRequired(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, addRoadmap)
	defer resetAddFlags()
	addPhase = ""
	if err := runRoadmapAdd(nil, []string{"new-widget"}); err == nil {
		t.Fatal("missing --phase must error")
	}
}

// TestRunRoadmapAdd_PropagatesError surfaces a package-level rejection.
func TestRunRoadmapAdd_PropagatesError(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, addRoadmap)
	defer resetAddFlags()
	addPhase = "Phase 1"
	if err := runRoadmapAdd(nil, []string{"auth-service"}); err == nil {
		t.Fatal("duplicate slug must propagate an error")
	}
}
