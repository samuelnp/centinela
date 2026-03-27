package docgen

import (
	"os"
	"strings"
	"testing"
)

func TestLoadRoadmapNodesFallbackAndHelpers(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                                                             //nolint:errcheck
	os.Chdir(d)                                                                                                   //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                                //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"features":[{"name":"a"},{"name":"b"}]}]}`), 0644) //nolint:errcheck
	nodes := loadRoadmapNodes()
	if len(nodes) != 2 || nodes[0].Name != "a" {
		t.Fatalf("unexpected fallback nodes: %#v", nodes)
	}
	if readFile("missing") != "" {
		t.Fatal("readFile missing should be empty")
	}
	if got := listFiles("*.none"); len(got) != 0 {
		t.Fatalf("expected no files, got %v", got)
	}
	os.WriteFile("specs.feature", []byte("Scenario: x\nScenario: y"), 0644) //nolint:errcheck
	if strings.Count(readFile("specs.feature"), "Scenario:") != 2 {
		t.Fatal("scenario counter fixture broken")
	}
}
