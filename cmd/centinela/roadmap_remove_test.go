package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// TestRunRoadmapRemove_Success deletes a planned, undepended feature.
func TestRunRoadmapRemove_Success(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, addRoadmap) // billing-api is planned, undepended
	if err := runRoadmapRemove(nil, []string{"billing-api"}); err != nil {
		t.Fatalf("runRoadmapRemove: %v", err)
	}
	r, err := roadmap.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, p := range r.Phases {
		for _, f := range p.Features {
			if f.Name == "billing-api" {
				t.Fatal("billing-api must be removed")
			}
		}
	}
}

// TestRunRoadmapRemove_Error surfaces a not-found rejection.
func TestRunRoadmapRemove_Error(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, addRoadmap)
	err := runRoadmapRemove(nil, []string{"ghost-feature"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("removing an absent feature must error, got %v", err)
	}
}
