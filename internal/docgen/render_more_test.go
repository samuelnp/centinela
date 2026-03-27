package docgen

import (
	"os"
	"strings"
	"testing"
)

func TestRenderHelpersAndGenerateErrors(t *testing.T) {
	if !strings.Contains(mermaidRoadmap([]RoadmapNode{{Name: "a"}, {Name: "b", DependsOn: []string{"a"}}}), "a --> b") {
		t.Fatal("roadmap mermaid missing edge")
	}
	if strings.Contains(mermaidEvidence([]EvidenceLink{{Role: "x", Handoff: ""}}), "-->") {
		t.Fatal("empty handoff should skip edge")
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
