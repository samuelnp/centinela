package docgen

import (
	"strings"
	"testing"
)

func TestRenderHelpersBranchCoverage(t *testing.T) {
	if got := firstLines("", 3); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
	if got := firstLines("a\nb", 0); got != "" {
		t.Fatalf("expected empty for zero lines, got %q", got)
	}
	if got := firstLines("a\nb", 3); got != "a\nb" {
		t.Fatalf("expected full content, got %q", got)
	}
	if !strings.Contains(roadmapList(nil), "No roadmap nodes") {
		t.Fatal("expected empty roadmap message")
	}
	r := roadmapList([]RoadmapNode{{Name: "f", DependsOn: []string{"a", "b"}}})
	if !strings.Contains(r, "a, b") {
		t.Fatal("expected dependency listing")
	}
	if got := cleanID("x y-z"); got != "x_y_z" {
		t.Fatalf("unexpected clean id: %q", got)
	}
	if got := cleanID("***"); got != "___" {
		t.Fatalf("expected underscore id, got %q", got)
	}
}

func TestGraphFallbackBranches(t *testing.T) {
	if !strings.Contains(mermaidRoadmap(nil), "No features detected") {
		t.Fatal("expected roadmap empty fallback")
	}
	r := mermaidRoadmap([]RoadmapNode{{Name: "feature-b", DependsOn: []string{"upstream feature"}}})
	if !strings.Contains(r, "upstream_feature") {
		t.Fatal("expected sanitized dependency id")
	}
	if !strings.Contains(mermaidSpecs(nil), "No specs found") {
		t.Fatal("expected specs empty fallback")
	}
}
