package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// resetReorderFlags restores the reorder command's globals after a test.
func resetReorderFlags() { reorderBefore, reorderAfter = "", "" }

const reorderCmdRoadmap = `{"phases":[{"name":"Phase 1","features":[` +
	`{"name":"auth-service"},{"name":"checkout-ui"},{"name":"billing-api"}]}]}`

// TestRunRoadmapReorder_Success repositions a feature within its phase.
func TestRunRoadmapReorder_Success(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, reorderCmdRoadmap)
	defer resetReorderFlags()
	reorderBefore = "auth-service"
	if err := runRoadmapReorder(nil, []string{"billing-api"}); err != nil {
		t.Fatalf("runRoadmapReorder: %v", err)
	}
	names := []string{}
	for _, p := range mustLoad(t).Phases {
		for _, f := range p.Features {
			names = append(names, f.Name)
		}
	}
	if len(names) == 0 || names[0] != "billing-api" {
		t.Fatalf("billing-api must move to the front: %v", names)
	}
}

// TestRunRoadmapReorder_RequiresAnchor errors when neither anchor is given.
func TestRunRoadmapReorder_RequiresAnchor(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, reorderCmdRoadmap)
	defer resetReorderFlags()
	if err := runRoadmapReorder(nil, []string{"billing-api"}); err == nil {
		t.Fatal("missing --before/--after must error")
	}
}

// TestRunRoadmapReorder_MutuallyExclusive rejects --before with --after.
func TestRunRoadmapReorder_MutuallyExclusive(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, reorderCmdRoadmap)
	defer resetReorderFlags()
	reorderBefore, reorderAfter = "auth-service", "checkout-ui"
	if err := runRoadmapReorder(nil, []string{"billing-api"}); err == nil {
		t.Fatal("--before and --after together must error")
	}
}

// TestRunRoadmapReorder_Error surfaces a package rejection (unknown slug).
func TestRunRoadmapReorder_Error(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, reorderCmdRoadmap)
	defer resetReorderFlags()
	reorderBefore = "auth-service"
	if err := runRoadmapReorder(nil, []string{"ghost"}); err == nil {
		t.Fatal("unknown slug must error")
	}
}
