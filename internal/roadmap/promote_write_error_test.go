package roadmap

import (
	"os"
	"testing"
)

// TestPromote_WriteRoadmapError returns partial error when writeRawRoadmap fails.
func TestPromote_WriteRoadmapError(t *testing.T) {
	setupPromoteDir(t)
	// Make the .workflow directory read-only so writeAtomic (rename) fails
	if os.Getenv("CI") != "" {
		t.Skip("file permission test skipped in CI")
	}
	scores, _ := ParseScores("9,9,9,9,9,9")
	// Replace roadmap.json with a directory to cause writeAtomic to fail on rename
	os.Remove(RoadmapFile) //nolint:errcheck
	if err := os.Mkdir(RoadmapFile, 0755); err != nil {
		t.Skip("cannot make roadmap.json a directory: " + err.Error())
	}
	// Now Promote will succeed up to writeRawRoadmap, then fail on rename
	if _, err := Promote(RoadmapFile, PromoteRequest{
		Slug: "hook-timeout-config", Phase: "Phase 5", Scores: scores,
	}); err == nil {
		t.Fatal("expected error when roadmap.json is a directory")
	}
}

// TestDefer_WriteRoadmapError returns error when writeRawRoadmap fails.
func TestDefer_WriteRoadmapError(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("file permission test skipped in CI")
	}
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	os.Chdir(d)                          //nolint:errcheck
	os.MkdirAll(".workflow", 0755)       //nolint:errcheck
	// write valid roadmap then turn it into a directory
	os.WriteFile(RoadmapFile, []byte(`{"phases":[]}`), 0644) //nolint:errcheck
	os.Remove(RoadmapFile)                                   //nolint:errcheck
	os.Mkdir(RoadmapFile, 0755)                              //nolint:errcheck
	if err := Defer(RoadmapFile, DeferOptions{Slug: "x", Summary: "s"}); err == nil {
		t.Fatal("expected error when roadmap.json is a directory")
	}
}
