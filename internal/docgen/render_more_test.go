package docgen

import (
	"os"
	"strings"
	"testing"
)

func TestRenderHelpersAndGenerateErrors(t *testing.T) {
	if !strings.Contains(mermaidRoadmap([]RoadmapNode{{Name: "a"}, {Name: "b", DependsOn: []string{"a"}}}), "-->") {
		t.Fatal("roadmap mermaid missing edge")
	}
	if !strings.Contains(mermaidSpecs([]string{"specs/alpha.feature"}), "Project Specs") {
		t.Fatal("spec map should include root")
	}
	tab := evidenceTable([]EvidenceLink{{Role: "r", Feature: "f", Step: "s", Outputs: []string{"a", "b"}}})
	if !strings.Contains(tab, "<br>") {
		t.Fatal("expected multi-output break")
	}
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	if err := Generate("", "x"); err == nil {
		t.Fatal("expected generate error on invalid inputs")
	}
}
